package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Token struct {
	typ   string
	name  string
	value any
	tail  string // everything after the closing ')' in the call: @ f(a) + 5 ;
}

var golang int
var func_spisok string
var typ string

func Token_length(tokens []Token) []Token {
	fmt.Print("Enter the .eye file name: ")
	var eye string
	fmt.Scan(&eye)
	eye += ".eye"
	file, err := os.ReadFile(eye)
	if err != nil {
		panic(err)
	}
	s := string(file)
	var st []string        // source text
	var kode_spisok string // source code from which whitespace is removed
	de := 0
	var slovo string
	for _, znak := range s {
		znaks := string(znak)
		if znaks == `"` && de == 0 { // check whether we are inside a string
			de++
		} else if de == 1 && znaks == `"` {
			de--
			slovo = `"` + slovo + `"`
			st = append(st, slovo)
			slovo = ""
		} else if de == 1 {
			slovo += znaks
		} else {
			if znaks != ";" {
				if znaks == "(" || znaks == ")" || znaks == "{" || znaks == "}" || znaks == "," {
					if kode_spisok != "" {
						st = append(st, kode_spisok)
						kode_spisok = ""
					}
					st = append(st, znaks)
				} else if znaks != " " && znaks != "\n" && znaks != "\t" && znaks != "\r" {
					kode_spisok += znaks
				} else {
					if kode_spisok != "" {
						st = append(st, kode_spisok)
						kode_spisok = ""
					}
				}
			} else {
				if kode_spisok != "" {
					st = append(st, kode_spisok)
					kode_spisok = ""
				}
				st = append(st, ";")
			}

		}

	}
	if kode_spisok != "" {
		st = append(st, kode_spisok)
	}
	kode := st
	var stroka []string
	depth := 0
	depth_krug := 0
	for _, v := range kode {
		var token Token
		if v == "end" {
			token.typ = "end"
			tokens = append(tokens, token)
			stroka = stroka[:0]
			continue
		} else if v == "{" {
			depth++
		} else if v == "}" {
			depth--
		} else if v == ")" {
			depth_krug--
		} else if v == "(" {
			depth_krug++
		}

		stroka = append(stroka, v)
		if v == ";" && depth == 0 && depth_krug == 0 {
			if len(stroka) > 0 && stroka[0] == "@" {
				if len(stroka) > 2 && stroka[2] == "(" {
					token.typ = "call" // required for function calls
					token.name = stroka[1]
					var args []string
					argDepth := 0 // nested parentheses: @ f(g(a), b)
					j := 3
					for ; j < len(stroka); j++ {
						if stroka[j] == "(" {
							argDepth++
						} else if stroka[j] == ")" {
							if argDepth == 0 {
								break
							}
							argDepth--
						}
						args = append(args, stroka[j])
					}
					token.value = strings.Join(args, " ")
					// trailing expression after the closing ')' (for example, "+ 5"), without the final ';'
					if j+1 < len(stroka) {
						rest := stroka[j+1:]
						if len(rest) > 0 && rest[len(rest)-1] == ";" {
							rest = rest[:len(rest)-1]
						}
						token.tail = strings.Join(rest, " ")
					}
				} else { // this branch creates function definitions
					token.typ = stroka[0]
					token.name = stroka[1]
					token.value = strings.Join(stroka[2:], " ")
				}

			} else if len(stroka) > 0 && (stroka[0] == "if" || stroka[0] == "for" || stroka[0] == "switch") {
				token.typ = stroka[0]
				token.name = strings.Join(stroka[1:], " ")
			} else {
				index_operation := 0
				for x, i := range stroka {
					y := x
					if x < len(stroka)-1 {
						y += 1
					}

					if strings.HasPrefix(i, "import") {
						z := x + 1
						if stroka[z] == "net" {
							golang += 1
							continue
						} else {
							token.typ = "import"
							token.name = stroka[z]
							token.value = ";"
						}
						continue
					}
					if strings.HasPrefix(i, "Int") || strings.HasPrefix(i, "Float") || strings.HasPrefix(i, "String") || strings.HasPrefix(i, "Bool") || strings.HasPrefix(i, "Double") || strings.HasPrefix(i, "File") || strings.HasPrefix(i, "net") {
						typ = i // remember the most recently used data type
						token.typ = i
						continue

					} else if strings.HasPrefix(token.typ, "net_") && x == 1 {
						// The net object and method name appear without '='.
						token.name = i
						if x+1 < len(stroka) && stroka[x+1] == "(" {
							var args []string
							parenDepth := 0
							for j := x + 2; j < len(stroka); j++ {
								if stroka[j] == "(" {
									parenDepth++
								}
								if stroka[j] == ")" {
									if parenDepth == 0 {
										break
									}
									parenDepth--
								}
								args = append(args, stroka[j])
							}
							token.value = strings.Join(args, " ")
							// The call has already been parsed: do not process the closing ')' again.
							break
						}
						continue
					} else if i == "^" {
						if typ == "" {
							log.Fatal("'^' used before a data type was declared")
						}
						token.typ = typ
						continue
					} else if (stroka[y] == "=" || stroka[y] == "+" || stroka[y] == "-" || stroka[y] == "/" || stroka[y] == "*" || stroka[y] == "%" || stroka[y] == "~") && index_operation == 0 {
						index_operation++
						token.name = i
					} else if i == "~" {
						if !strings.HasPrefix(token.typ, "net_") || token.value == nil {
							token.value = i
						}
					} else if i == "=" {
						// We found '='; the name is already set. Collect the value AFTER it.
						var value_name string
						for j := x + 1; j < len(stroka)-1; j++ { // x+1 skips '='
							value_name += stroka[j]
						}
						token.value = value_name
						break // Stop the loop; the value has been found
					} else {
					}
				}

			}
			tokens = append(tokens, token)
			stroka = stroka[:0]
		}

	}
	return tokens
}

func sozdanie(tokens []Token, file *os.File, file2 *os.File) {
	func_spisok = "" // FIXED: reset on every new run
	proverka_import := true
	// true = the user explicitly declared the variable (String message.Start = ...)
	declared := func(n string) bool {
		for _, t := range tokens {
			if t.typ != "call" && strings.HasPrefix(t.name, n+".") {
				return true
			}
		}
		return false
	}
	userMessage := declared("message")
	userAnswer := declared("answer")
	_, er := file.WriteString("#include <stdio.h>\n#include <math.h>\n#include <stdbool.h>\n#include <ctype.h>\n#include <stdlib.h>\n#include <string.h>\n\ntypedef struct Bool {\n    _Bool value;\n} Bool;\n\ntypedef struct Float {\n    float value;\n    int(*Conv_int_min)(struct Float);\n    int(*Conv_int_max)(struct Float);\n    char*(*Float_to_string)(struct Float, char[50]);\n    int(*Max_float)(struct Float, float);\n} Float;\n\nint Conv_int_min(Float value) { return (int)value.value; }\nint Conv_int_max(Float value) { return (int)roundf(value.value); }\n\nchar* Float_to_string(Float value, char s[50]) {\n    sprintf(s, \"%f\", value.value);\n    return s;\n}\n\nint Max_float(Float value, float value2) {\n    if (value.value > value2) return 1;\n    if (value2 == value.value) return 0;\n    return -1;\n}\n\ntypedef struct Int {\n    int value;\n    float(*Float_conv)(struct Int);\n    int(*Max_int)(struct Int, int);\n    char*(*Int_to_string)(struct Int, char[50]);\n    int(*Repeat_int)(struct Int, int);\n    int(*Chet)(struct Int);\n    int(*Int_length)(struct Int);\n} Int;\n\nfloat Int_to_float(Int value) { return (float)value.value; }\n\nint Max_int(Int value, int value2) {\n    if (value.value > value2) return 1;\n    if (value2 == value.value) return 0;\n    return -1;\n}\n\nchar* Int_to_string(Int value, char s[50]) {\n    sprintf(s, \"%d\", value.value);\n    return s;\n}\n\nint Repeat_int(Int value, int exp) {\n    int x = 1;\n    for (int i = 0; i < exp; i++) x *= value.value;\n    return x;\n}\n\nint Chet(Int value) { return value.value % 2 == 0; }\n\nint Int_length(Int value) {\n    if (value.value == 0) return 1;\n    int count = 0;\n    int tmp = value.value;\n    if (tmp < 0) tmp = -tmp;\n    while (tmp > 0) { tmp /= 10; count++; }\n    return count;\n}\n\ntypedef struct Double {\n    double value;\n    int(*Conv_int_min)(struct Double);\n    int(*Conv_int_max)(struct Double);\n    char*(*String_double)(struct Double, char[50]);\n    int(*Max_double)(struct Double, double);\n} Double;\n\nint Conv_int_min_double(Double value) { return (int)value.value; }\nint Conv_int_max_double(Double value) { return (int)round(value.value); }\n\nchar* Double_to_string(Double value, char s[50]) {\n    sprintf(s, \"%f\", value.value);\n    return s;\n}\n\nint Max_double(Double value, double value2) {\n    if (value.value > value2) return 1;\n    if (value2 == value.value) return 0;\n    return -1;\n}\n\ntypedef struct String {\n    void(*Size_stirng)(struct String*, int);\n    char* value;\n    int(*String_to_int)(struct String);\n    float(*String_to_float)(struct String);\n    char (*Repeat_string)(struct String, int);\n    void(*Lower)(struct String*);\n    void(*Uper)(struct String*);\n    int(*Contains)(struct String, char*);\n} String;\n\nint Contains(String value, char* value2) {\n    if (value.value == NULL) return 0;\n    return strstr(value.value, value2) != NULL;\n}\n\nint String_to_int(String value) { return atoi(value.value); }\nfloat String_to_float(String value) { return (float)atof(value.value); }\n\nvoid Size_string(String* value, int size) {\n    if (size <= 0) return;\n    char* v = malloc(size);\n    if (v == NULL) return;\n    v[0] = '\\0';\n    if (value->value != NULL) {\n        size_t oldlen = strlen(value->value);\n        size_t copylen = (oldlen < (size_t)(size - 1)) ? oldlen : (size_t)(size - 1);\n        memcpy(v, value->value, copylen);\n        v[copylen] = '\\0';\n    }\n    value->value = v;\n}\n\nvoid Upper(String* value) {\n    if (value->value == NULL) return;\n    for (int i = 0; value->value[i] != '\\0'; i++) {\n        value->value[i] = (char)toupper((unsigned char)value->value[i]);\n    }\n}\n\nvoid Lower(String* value) {\n    if (value->value == NULL) return;\n    for (int i = 0; value->value[i] != '\\0'; i++) {\n        value->value[i] = (char)tolower((unsigned char)value->value[i]);\n    }\n}\n\ntypedef struct File {\n    char* value;\n    char name[100];\n    int(*Write_file)(struct File*);\n    int(*Read_file)(struct File*);\n    long int(*Size_info)(struct File*);\n    void(*append)(struct File*, char[]);\n} File;\n\nint Write_file(File* value) {\n    FILE* f = fopen(value->name, \"w\");\n    if (f == NULL) return -1;\n    int col = fprintf(f, \"%s\", value->value);\n    fclose(f);\n    return col;\n}\n\nint Read_file(File* value) {\n    FILE* f = fopen(value->name, \"r\");\n    if (!f) return 1;\n    long size = value->Size_info(value);\n    if (size <= 0) { fclose(f); return 1; }\n    value->value = malloc(size + 1);\n    if (value->value == NULL) { fclose(f); return 1; }\n    if (fgets(value->value, (int)(size + 1), f) != NULL) {\n        fclose(f);\n        return 0;\n    }\n    fclose(f);\n    return 1;\n}\n\nvoid append(File* value, char value2[]) {\n    FILE* f = fopen(value->name, \"a\");\n    if (f == NULL) return;\n    fputs(value2, f);\n    fclose(f);\n}\n\nlong int Size_info(File* value) {\n    FILE* f = fopen(value->name, \"rb\");\n    if (!f) return -1;\n    fseek(f, 0, SEEK_END);\n    long size = ftell(f);\n    fclose(f);\n    return size;\n}\n\ntypedef struct String_list {\n    String* value;\n    void(*Delete_string)(struct String_list*);\n    void(*Nul_stirng)(struct String_list*, int);\n    void(*Redect_string)(struct String_list*, int);\n    String(*Join_string)(struct String_list*);\n    int Size;\n} String_list;\n\nvoid Delete_string(String_list* value) {\n    if (value->value != NULL) {\n        free(value->value);\n        value->value = NULL;\n    }\n    value->Size = 0;\n}\n\nvoid Nul_string(String_list* value, int caunt) {\n    if (value->value != NULL) free(value->value);\n    value->value = calloc(caunt, sizeof(String));\n    if (value->value == NULL) { value->Size = 0; return; }\n    value->Size = caunt;\n}\n\nvoid Redirect_string(String_list* value, int caunt) {\n    String* tmp = realloc(value->value, caunt * sizeof(String));\n    if (tmp != NULL) { value->value = tmp; value->Size = caunt; }\n}\n\nString Join_string(String_list* value) {\n    String result;\n    result.value = NULL;\n    if (value->value == NULL || value->Size <= 0) {\n        result.value = malloc(1);\n        if (result.value) result.value[0] = '\\0';\n        return result;\n    }\n    size_t total = 1;\n    for (int i = 0; i < value->Size; i++) {\n        if (value->value[i].value) total += strlen(value->value[i].value) + 1;\n    }\n    result.value = malloc(total);\n    if (result.value == NULL) return result;\n    result.value[0] = '\\0';\n    for (int i = 0; i < value->Size; i++) {\n        if (value->value[i].value) strcat(result.value, value->value[i].value);\n        if (i < value->Size - 1) strcat(result.value, \",\");\n    }\n    return result;\n}\n\ntypedef unsigned char Norm;\n\ntypedef struct Byte {\n    Norm* value;\n    void(*Delete_byte)(struct Byte*);\n    void(*Nul_byte)(struct Byte*, int);\n    void(*Redect_byte)(struct Byte*, int);\n    int Size;\n} Byte;\n\nvoid Delete_byte(Byte* value) {\n    if (value->value != NULL) {\n        free(value->value);\n        value->value = NULL;\n    }\n    value->Size = 0;\n}\n\nvoid Nul_byte(Byte* value, int caunt) {\n    if (value->value != NULL) free(value->value);\n    value->value = calloc(caunt, sizeof(Norm));\n    if (value->value == NULL) { value->Size = 0; return; }\n    value->Size = caunt;\n}\n\nvoid Redirect_byte(Byte* value, int caunt) {\n    Norm* tmp = realloc(value->value, caunt * sizeof(Norm));\n    if (tmp != NULL) { value->value = tmp; value->Size = caunt; }\n}\n\ntypedef struct Int_list {\n    Int* value;\n    void(*Delete_int)(struct Int_list*);\n    void(*Nul_int)(struct Int_list*, int);\n    void(*Redect_int)(struct Int_list*, int);\n    String(*Join_int)(struct Int_list*);\n    int Size;\n} Int_list;\n\nvoid Delete_int(Int_list* value) {\n    if (value->value != NULL) { free(value->value); value->value = NULL; }\n    value->Size = 0;\n}\n\nvoid Nul_int(Int_list* value, int caunt) {\n    if (value->value != NULL) free(value->value);\n    value->value = calloc(caunt, sizeof(Int));\n    if (value->value == NULL) { value->Size = 0; return; }\n    value->Size = caunt;\n}\n\nvoid Redirect_int(Int_list* value, int caunt) {\n    Int* tmp = realloc(value->value, caunt * sizeof(Int));\n    if (tmp != NULL) { value->value = tmp; value->Size = caunt; }\n}\n\nString Join_int(Int_list* value) {\n    String result;\n    result.value = NULL;\n    if (value->value == NULL || value->Size <= 0) {\n        result.value = malloc(1);\n        if (result.value) result.value[0] = '\\0';\n        return result;\n    }\n    size_t total = 1;\n    char tmp[32];\n    for (int i = 0; i < value->Size; i++) {\n        total += snprintf(tmp, sizeof(tmp), \"%d\", value->value[i].value) + 1;\n    }\n    result.value = malloc(total);\n    if (result.value == NULL) return result;\n    result.value[0] = '\\0';\n    for (int i = 0; i < value->Size; i++) {\n        char tmp2[32];\n        snprintf(tmp2, sizeof(tmp2), \"%d\", value->value[i].value);\n        strcat(result.value, tmp2);\n        if (i < value->Size - 1) strcat(result.value, \",\");\n    }\n    return result;\n}\n\ntypedef struct Float_list {\n    Float* value;\n    void(*Delete_float)(struct Float_list*);\n    void(*Nul_float)(struct Float_list*, int);\n    void(*Redect_float)(struct Float_list*, int);\n    String(*Join_float)(struct Float_list*);\n    int Size;\n} Float_list;\n\nvoid Delete_float(Float_list* value) {\n    if (value->value != NULL) { free(value->value); value->value = NULL; }\n    value->Size = 0;\n}\n\nvoid Nul_float(Float_list* value, int caunt) {\n    if (value->value != NULL) free(value->value);\n    value->value = calloc(caunt, sizeof(Float));\n    if (value->value == NULL) { value->Size = 0; return; }\n    value->Size = caunt;\n}\n\nvoid Redirect_float(Float_list* value, int caunt) {\n    Float* tmp = realloc(value->value, caunt * sizeof(Float));\n    if (tmp != NULL) { value->value = tmp; value->Size = caunt; }\n}\n\nString Join_float(Float_list* value) {\n    String result;\n    result.value = NULL;\n    if (value->value == NULL || value->Size <= 0) {\n        result.value = malloc(1);\n        if (result.value) result.value[0] = '\\0';\n        return result;\n    }\n    size_t total = 1;\n    char tmp[64];\n    for (int i = 0; i < value->Size; i++) {\n        total += snprintf(tmp, sizeof(tmp), \"%f\", value->value[i].value) + 1;\n    }\n    result.value = malloc(total);\n    if (result.value == NULL) return result;\n    result.value[0] = '\\0';\n    for (int i = 0; i < value->Size; i++) {\n        char tmp2[64];\n        snprintf(tmp2, sizeof(tmp2), \"%f\", value->value[i].value);\n        strcat(result.value, tmp2);\n        if (i < value->Size - 1) strcat(result.value, \",\");\n    }\n    return result;\n}\n\ntypedef struct Double_list {\n    Double* value;\n    void(*Delete_double)(struct Double_list*);\n    void(*Nul_double)(struct Double_list*, int);\n    void(*Redect_double)(struct Double_list*, int);\n    String(*Join_double)(struct Double_list*);\n    int Size;\n} Double_list;\n\nvoid Delete_double(Double_list* value) {\n    if (value->value != NULL) { free(value->value); value->value = NULL; }\n    value->Size = 0;\n}\n\nvoid Nul_double(Double_list* value, int caunt) {\n    if (value->value != NULL) free(value->value);\n    value->value = calloc(caunt, sizeof(Double));\n    if (value->value == NULL) { value->Size = 0; return; }\n    value->Size = caunt;\n}\n\nvoid Redirect_double(Double_list* value, int caunt) {\n    Double* tmp = realloc(value->value, caunt * sizeof(Double));\n    if (tmp != NULL) { value->value = tmp; value->Size = caunt; }\n}\n\nString Join_double(Double_list* value) {\n    String result;\n    result.value = NULL;\n    if (value->value == NULL || value->Size <= 0) {\n        result.value = malloc(1);\n        if (result.value) result.value[0] = '\\0';\n        return result;\n    }\n    size_t total = 1;\n    char tmp[64];\n    for (int i = 0; i < value->Size; i++) {\n        total += snprintf(tmp, sizeof(tmp), \"%f\", value->value[i].value) + 1;\n    }\n    result.value = malloc(total);\n    if (result.value == NULL) return result;\n    result.value[0] = '\\0';\n    for (int i = 0; i < value->Size; i++) {\n        char tmp2[64];\n        snprintf(tmp2, sizeof(tmp2), \"%f\", value->value[i].value);\n        strcat(result.value, tmp2);\n        if (i < value->Size - 1) strcat(result.value, \",\");\n    }\n    return result;\n}\n\ntypedef struct Bool_list {\n    Bool* value;\n    void(*Delete_bool)(struct Bool_list*);\n    void(*Nul_bool)(struct Bool_list*, int);\n    void(*Redect_bool)(struct Bool_list*, int);\n    String(*Join_bool)(struct Bool_list*);\n    int Size;\n} Bool_list;\n\nvoid Delete_bool(Bool_list* value) {\n    if (value->value != NULL) { free(value->value); value->value = NULL; }\n    value->Size = 0;\n}\n\nvoid Nul_bool(Bool_list* value, int caunt) {\n    if (value->value != NULL) free(value->value);\n    value->value = calloc(caunt, sizeof(Bool));\n    if (value->value == NULL) { value->Size = 0; return; }\n    value->Size = caunt;\n}\n\nvoid Redirect_bool(Bool_list* value, int caunt) {\n    Bool* tmp = realloc(value->value, caunt * sizeof(Bool));\n    if (tmp != NULL) { value->value = tmp; value->Size = caunt; }\n}\n\nString Join_bool(Bool_list* value) {\n    String result;\n    result.value = NULL;\n    if (value->value == NULL || value->Size <= 0) {\n        result.value = malloc(1);\n        if (result.value) result.value[0] = '\\0';\n        return result;\n    }\n    size_t total = 1;\n    for (int i = 0; i < value->Size; i++) total += 2;\n    result.value = malloc(total);\n    if (result.value == NULL) return result;\n    result.value[0] = '\\0';\n    for (int i = 0; i < value->Size; i++) {\n        strcat(result.value, value->value[i].value ? \"1\" : \"0\");\n        if (i < value->Size - 1) strcat(result.value, \",\");\n    }\n    return result;\n}\n\ntypedef struct File_list {\n    File* value;\n    void(*Delete_file)(struct File_list*);\n    void(*Nul_file)(struct File_list*, int);\n    void(*Redect_file)(struct File_list*, int);\n    String(*Join_file)(struct File_list*);\n    int Size;\n} File_list;\n\nvoid Delete_file(File_list* value) {\n    if (value->value != NULL) { free(value->value); value->value = NULL; }\n    value->Size = 0;\n}\n\nvoid Nul_file(File_list* value, int caunt) {\n    if (value->value != NULL) free(value->value);\n    value->value = calloc(caunt, sizeof(File));\n    if (value->value == NULL) { value->Size = 0; return; }\n    value->Size = caunt;\n}\n\nvoid Redirect_file(File_list* value, int caunt) {\n    File* tmp = realloc(value->value, caunt * sizeof(File));\n    if (tmp != NULL) { value->value = tmp; value->Size = caunt; }\n}\n\nString Join_file(File_list* value) {\n    String result;\n    result.value = NULL;\n    if (value->value == NULL || value->Size <= 0) {\n        result.value = malloc(1);\n        if (result.value) result.value[0] = '\\0';\n        return result;\n    }\n    size_t total = 1;\n    for (int i = 0; i < value->Size; i++) total += strlen(value->value[i].name) + 1;\n    result.value = malloc(total);\n    if (result.value == NULL) return result;\n    result.value[0] = '\\0';\n    for (int i = 0; i < value->Size; i++) {\n        strcat(result.value, value->value[i].name);\n        if (i < value->Size - 1) strcat(result.value, \",\");\n    }\n    return result;\n}\n\n")
	if er != nil {
		log.Fatal(er)
	}
	if golang == 1 {
		_, ers := file2.WriteString("package main\n\nimport (\n\t\"bytes\"\n\t\"fmt\"\n\t\"io\"\n\t\"log\"\n\t\"net\"\n\t\"net/http\"\n\t\"strconv\"\n\n\t\"strings\"\n\t\"sync\"\n\t\"time\"\n)\n\nvar (\n\tconn net.Conn\n\tmu   sync.Mutex\n)\n\nfunc Init() {\n\tvar err error\n\tconn, err = net.Dial(\"unix\", \"/tmp/eye.sock\")\n\tif err != nil {\n\t\tpanic(err)\n\t}\n}\n\nfunc Uds(s string) string {\n\tmu.Lock()\n\tdefer mu.Unlock()\n\n\tconn.SetReadDeadline(time.Now().Add(5 * time.Second))\n\n\tif _, er := conn.Write([]byte(s + \"\\n\")); er != nil {\n\t\tlog.Println(er)\n\t\treturn \"\"\n\t}\n\n\tvar result strings.Builder\n\tbuf := make([]byte, 4096)\n\tfor {\n\t\tcol, er := conn.Read(buf)\n\t\tif col > 0 {\n\t\t\tresult.Write(buf[:col])\n\t\t\tif buf[col-1] == '\\n' {\n\t\t\t\tbreak\n\t\t\t}\n\t\t}\n\t\tif er != nil {\n\t\t\tif er != io.EOF {\n\t\t\t\tlog.Println(er)\n\t\t\t}\n\t\t\tbreak\n\t\t}\n\t}\n\treturn strings.TrimSpace(result.String())\n}\n\nfunc Address(r *http.Request) {\n\ts := fmt.Sprintf(\"%s:%v\", r.URL, r.RemoteAddr)\n\tv := Uds(s)\n\tif v != \"ok\" {\n\t\tv = Uds(s)\n\t\tif v != \"ok\" {\n\t\t\tlog.Println(\"ERROR:\", r.URL)\n\t\t}\n\t}\n}\n\nfunc Query(r *http.Request) {\n\ts := fmt.Sprintf(\"%s:%v\", r.URL, r.URL.Query())\n\tv := Uds(s)\n\tif v != \"ok\" {\n\t\tv = Uds(s)\n\t\tif v != \"ok\" {\n\t\t\tlog.Println(\"ERROR:\", r.URL)\n\t\t}\n\t}\n}\n\nfunc Method(r *http.Request) {\n\ts := fmt.Sprintf(\"%s:%v\", r.URL, r.Method)\n\tv := Uds(s)\n\tif v != \"ok\" {\n\t\tv = Uds(s)\n\t\tif v != \"ok\" {\n\t\t\tlog.Println(\"ERROR:\", r.URL)\n\t\t}\n\t}\n}\n\nfunc Header(r *http.Request) {\n\ts := fmt.Sprintf(\"%s:%v\", r.URL, r.Header)\n\tv := Uds(s)\n\tif v != \"ok\" {\n\t\tv = Uds(s)\n\t\tif v != \"ok\" {\n\t\t\tlog.Println(\"ERROR:\", r.URL)\n\t\t}\n\t}\n}\n\nfunc info_user(r *http.Request) {\n\ts := fmt.Sprintf(\"%s:%v\", r.URL, r.Referer())\n\tv := Uds(s)\n\tif v != \"ok\" {\n\t\tv = Uds(s)\n\t\tif v != \"ok\" {\n\t\t\tlog.Println(\"ERROR:\", r.URL)\n\t\t}\n\t}\n}\n\nfunc Form(r *http.Request) {\n\ts := fmt.Sprintf(\"%s:%v\", r.URL, r.Form)\n\tv := Uds(s)\n\tif v != \"ok\" {\n\t\tv = Uds(s)\n\t\tif v != \"ok\" {\n\t\t\tlog.Println(\"ERROR:\", r.URL)\n\t\t}\n\t}\n}\n\nfunc Byte(r *http.Request) {\n\tbody, err := io.ReadAll(r.Body)\n\tif err != nil {\n\t\tlog.Println(\"ERROR READ BODY:\", err)\n\t\treturn\n\t}\n\tr.Body.Close()\n\ts := fmt.Sprintf(\"%s:%s\", r.URL, string(body))\n\tv := Uds(s)\n\tif v != \"ok\" {\n\t\tv = Uds(s)\n\t\tif v != \"ok\" {\n\t\t\tlog.Println(\"ERROR:\", r.URL)\n\t\t}\n\t}\n}\n\nfunc Cookie(r *http.Request) {\n\ts := fmt.Sprintf(\"%s:%v\", r.URL, r.Cookies())\n\tv := Uds(s)\n\tif v != \"ok\" {\n\t\tv = Uds(s)\n\t\tif v != \"ok\" {\n\t\t\tlog.Println(\"ERROR:\", r.URL)\n\t\t}\n\t}\n}\n\nfunc Write(w http.ResponseWriter, r *http.Request) {\n\ts := fmt.Sprintf(\"!value %s\", r.URL)\n\tv := Uds(s)\n\ttyp, value, ok := strings.Cut(v, \"!\")\n\tif !ok {\n\t\tlog.Println(\"ERROR WRITE: \", r.URL)\n\t\treturn\n\t}\n\tw.Header().Set(\"Content-Type\", typ)\n\tfmt.Fprint(w, value)\n}\n\ntype net_http_server struct {\n\tAddress   func(*http.Request)\n\tQuery     func(*http.Request)\n\tMethod    func(*http.Request)\n\tHeader    func(*http.Request)\n\tValueUser func(*http.Request)\n\tForm      func(*http.Request)\n\tByte      func(*http.Request)\n\tCookie    func(*http.Request)\n\tWrite     func(http.ResponseWriter, *http.Request)\n}\n\nvar httpClient = &http.Client{Timeout: 15 * time.Second}\n\nfunc HttpGet(url string) string {\n\tresp, err := httpClient.Get(url)\n\tif err != nil {\n\t\treturn \"error!\" + err.Error()\n\t}\n\tdefer resp.Body.Close()\n\tbody, _ := io.ReadAll(resp.Body)\n\treturn \"ok!\" + resp.Header.Get(\"Content-Type\") + \"!\" + strconv.Itoa(resp.StatusCode) + \"!\" + string(body)\n}\n\nfunc HttpWrite(method, value, url string) string {\n\tvar reader io.Reader\n\tif value != \"\" {\n\t\treader = strings.NewReader(value)\n\t}\n\treq, err := http.NewRequest(strings.ToUpper(method), url, reader)\n\tif err != nil {\n\t\treturn \"error!\" + err.Error()\n\t}\n\tresp, err := httpClient.Do(req)\n\tif err != nil {\n\t\treturn \"error!\" + err.Error()\n\t}\n\tdefer resp.Body.Close()\n\tdata, _ := io.ReadAll(resp.Body)\n\treturn \"ok!\" + resp.Header.Get(\"Content-Type\") + \"!\" + strconv.Itoa(resp.StatusCode) + \"!\" + string(data)\n}\n\nfunc HttpUpload(url string, data []byte) string {\n\treq, err := http.NewRequest(\"POST\", url, bytes.NewReader(data))\n\tif err != nil {\n\t\treturn \"error!\" + err.Error()\n\t}\n\treq.Header.Set(\"Content-Type\", \"application/octet-stream\")\n\tresp, err := httpClient.Do(req)\n\tif err != nil {\n\t\treturn \"error!\" + err.Error()\n\t}\n\tdefer resp.Body.Close()\n\tbody, _ := io.ReadAll(resp.Body)\n\treturn \"ok!\" + resp.Header.Get(\"Content-Type\") + \"!\" + strconv.Itoa(resp.StatusCode) + \"!\" + string(body)\n}\n\ntype net_http_client struct {\n\tHttpGet    func(url string) string\n\tHttpWrite  func(method, value, url string) string\n\tHttpUpload func(url string, data []byte) string\n}\n\nconst BlockSize = 1024\n\nfunc StartTCP(address string, blocks int) error {\n\tif blocks <= 0 {\n\t\treturn fmt.Errorf(\"blocks must be greater than 0\")\n\t}\n\n\tmaxBytes := blocks * BlockSize\n\n\tlistener, err := net.Listen(\"tcp\", address)\n\tif err != nil {\n\t\treturn err\n\t}\n\tdefer listener.Close()\n\n\tfor {\n\t\tcon, err := listener.Accept()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\n\t\tgo func(con net.Conn) {\n\t\t\tdefer con.Close()\n\n\t\t\tcon.SetReadDeadline(time.Now().Add(5 * time.Second))\n\n\t\t\tbuf := make([]byte, maxBytes)\n\t\t\tn, err := io.ReadFull(con, buf)\n\n\t\t\tif err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {\n\t\t\t\treturn\n\t\t\t}\n\n\t\t\tpayload := strings.TrimRight(string(buf[:n]), \"\\r\\n\")\n\t\t\tfinal := Uds(fmt.Sprintf(\"?%s:tcp:%s\", address, payload))\n\t\t\t_, _ = con.Write([]byte(final))\n\t\t}(con)\n\t}\n}\n\nfunc StartUDP(address string, blocks int) error {\n\tif blocks <= 0 {\n\t\treturn fmt.Errorf(\"blocks must be greater than 0\")\n\t}\n\n\tmaxBytes := blocks * BlockSize\n\n\tcon, err := net.ListenPacket(\"udp\", address)\n\tif err != nil {\n\t\treturn err\n\t}\n\tdefer con.Close()\n\n\tbuffer := make([]byte, maxBytes)\n\n\tfor {\n\t\tn, clientAddress, err := con.ReadFrom(buffer)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t\tpayload := strings.TrimRight(string(buffer[:n]), \"\\r\\n\")\n\t\tfinal := Uds(fmt.Sprint(\"?\", address, \":\", \"udp:\", payload))\n\t\t_, err = con.WriteTo([]byte(final), clientAddress)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t}\n}\n\ntype net_prtotocol_server struct {\n\tStartTCP func(string, int)\n\tStartUDP func(string, int)\n}\n\nfunc TCP_user(address string) {\n\tcon, err := net.Dial(\"tcp\", address)\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n\tdefer con.Close()\n\t_ = con.SetReadDeadline(time.Now().Add(3 * time.Second))\n\t_, err = con.Write([]byte(Uds(fmt.Sprintf(\"?tcp:%s:user\", address))))\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n\n\tif tcpCon, ok := con.(*net.TCPConn); ok {\n\t\t_ = tcpCon.CloseWrite()\n\t}\n\n\tanswer := make([]byte, 1024)\n\n\tn, err := con.Read(answer)\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n\tUds(fmt.Sprint(\"~\", address, \":\", string(answer[:n])))\n\n}\n\nfunc UDP_user(address string) {\n\tcon, err := net.Dial(\"udp\", address)\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n\tdefer con.Close()\n\n\t_, err = con.Write([]byte(Uds(fmt.Sprintf(\"?udp:%s:user\", address))))\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n\n\t_ = con.SetReadDeadline(time.Now().Add(3 * time.Second))\n\n\tanswer := make([]byte, 1024)\n\n\tn, err := con.Read(answer)\n\tif err != nil {\n\t\tlog.Fatal(err)\n\t}\n\tUds(fmt.Sprint(\"~\", address, \":\", string(answer[:n])))\n\n}\n\ntype net_prototcol_user struct {\n\tUDP_user func(string)\n\tTCP_user func(string)\n}\n\nfunc main() {\n\tInit()\n\tdefer conn.Close()\n\t")
		if ers != nil {
			log.Fatal(ers)
		}
	}
	for _, v := range tokens {
		if v.typ == "@" {
			func_spisok += fmt.Sprintf("%v %v\n", v.name, v.value)
		}
	}
	_, erq := file.WriteString(func_spisok + "\n")
	if erq != nil {
		log.Fatal(erq)
	}

	for _, v := range tokens {
		if v.typ == "import" || proverka_import { // check whether the import section has ended
			if v.name == "stop_import" {
				proverka_import = false
				_, errf := file.WriteString("int main(void){\n")
				if errf != nil {
					log.Fatal(errf)
				}
				// In net mode: message (incoming string from Go) and answer (Go response)
				// are implicit unless the user declares them explicitly.
				if golang == 1 {
					if !userMessage {
						_, errf = file.WriteString("char message_buf[4096];\nString message;\nmessage.value = message_buf;\nif (fgets(message_buf, sizeof(message_buf), stdin) == NULL) message_buf[0] = '\\0';\nmessage_buf[strcspn(message_buf, \"\\r\\n\")] = '\\0';\n")
						if errf != nil {
							log.Fatal(errf)
						}
					}
					if !userAnswer {
						_, errf = file.WriteString("char answer_buf[4096];\nString answer;\nanswer.value = answer_buf;\nanswer_buf[0] = '\\0';\n")
						if errf != nil {
							log.Fatal(errf)
						}
					}
				}
				continue
			}
			if v.typ == "import" {
				fil, ers := os.ReadFile(v.name)
				if ers != nil {
					log.Fatal("import error ")
				}
				_, errf := file.Write(fil)
				if errf != nil {
					log.Fatal(errf)
				}
				_, errs := file.WriteString("\n")
				if errs != nil {
					log.Fatal(errs)
				}
			}

		}

		if v.typ == "@" {
			continue
		}
		if v.typ == "call" {
			tail := ""
			if v.tail != "" {
				tail = " " + v.tail
			}
			_, errf := file.WriteString(fmt.Sprintf("%v (%v)%v;\n", v.name, v.value, tail))
			if errf != nil {
				log.Fatal(errf)
			}
		} else if v.typ == "if" || v.typ == "for" || v.typ == "switch" {
			_, errf := file.WriteString(fmt.Sprintf("%v %v\n", v.typ, v.name))
			if errf != nil {
				log.Fatal(errf)
			}
		} else if v.typ == "end" {
			if golang == 1 && !userAnswer {
				// the response is sent back to Go: a string prefixed with '$'
				_, errf := file.WriteString("printf(\"$%s\\n\", answer.value);\nfflush(stdout);\n")
				if errf != nil {
					log.Fatal(errf)
				}
			}
			_, errf := file.WriteString("return 0;\n}\n")
			if errf != nil {
				log.Fatal(errf)
			}
			if golang == 1 {
				_, ers := file2.WriteString("}")
				if ers != nil {
					log.Fatal(ers)
				}

			}
		} else if v.typ == "net_http_server" {
			if strings.HasSuffix(v.name, ".Start") {
				// The route already contains quotation marks; do not add them again.
				_, ers := file2.WriteString(fmt.Sprintf("http.HandleFunc(%s, func(w http.ResponseWriter, r *http.Request) {\n", v.value))
				if ers != nil {
					log.Fatal(ers)
				}
			} else if strings.HasSuffix(v.name, ".Stop") {
				_, ers := file2.WriteString("}")
				if ers != nil {
					log.Fatal(ers)
				}
			} else if strings.HasSuffix(v.name, ".Where") {
				_, ers := file2.WriteString(")\n")
				if ers != nil {
					log.Fatal(ers)
				}
			} else if strings.HasSuffix(v.name, ".Listen") {
				_, ers := file2.WriteString(fmt.Sprintf("log.Fatal(http.ListenAndServe(%s, nil))\n", v.value))
				if ers != nil {
					log.Fatal(ers)
				}
			} else {
				methodCall := ""
				switch {
				case strings.HasSuffix(v.name, ".Address"):
					methodCall = "Address(r)\n"
				case strings.HasSuffix(v.name, ".Query"):
					methodCall = "Query(r)\n"
				case strings.HasSuffix(v.name, ".Method"):
					methodCall = "Method(r)\n"
				case strings.HasSuffix(v.name, ".Header"):
					methodCall = "Header(r)\n"
				case strings.HasSuffix(v.name, ".ValueUser"):
					methodCall = "info_user(r)\n"
				case strings.HasSuffix(v.name, ".Form"):
					methodCall = "Form(r)\n"
				case strings.HasSuffix(v.name, ".Byte"):
					methodCall = "Byte(r)\n"
				case strings.HasSuffix(v.name, ".Cookie"):
					methodCall = "Cookie(r)\n"
				case strings.HasSuffix(v.name, ".Write"):
					methodCall = "Write(w, r)\n"
				default:
					methodCall = v.name + "\n"
				}
				_, ers := file2.WriteString(methodCall)
				if ers != nil {
					log.Fatal(ers)
				}
			}
		} else if strings.HasPrefix(v.typ, "net_") {
			// The other net types are also translated into ready-to-use functions in the Go package.
			methodCall := ""
			switch {
			case strings.HasSuffix(v.name, ".StartTCP"):
				methodCall = "StartTCP"
			case strings.HasSuffix(v.name, ".StartUDP"):
				methodCall = "StartUDP"
			case strings.HasSuffix(v.name, ".TCP_user"):
				methodCall = "TCP_user"
			case strings.HasSuffix(v.name, ".UDP_user"):
				methodCall = "UDP_user"
			case strings.HasSuffix(v.name, ".HttpGet"):
				methodCall = "HttpGet"
			case strings.HasSuffix(v.name, ".HttpWrite"):
				methodCall = "HttpWrite"
			case strings.HasSuffix(v.name, ".HttpUpload"):
				methodCall = "HttpUpload"
			}
			if methodCall != "" {
				_, ers := file2.WriteString(fmt.Sprintf("%s(%v)\n", methodCall, v.value))
				if ers != nil {
					log.Fatal(ers)
				}
			} else if v.name != "" {
				_, ers := file2.WriteString(v.name + "\n")
				if ers != nil {
					log.Fatal(ers)
				}
			}
		} else if strings.HasSuffix(v.name, ".Start") { // C and Go cannot call the method directly, so we call it first and then create .Value
			name, _, ok := strings.Cut(v.name, ".")
			if !ok {
				log.Fatal("error in .Start")
			}
			if golang == 1 && strings.HasPrefix(v.typ, "net") {
				_, ers := file2.WriteString(fmt.Sprintf(" var %s %s\n", name, v.typ))
				if ers != nil {
					log.Fatal(ers)
				}
			} else if golang > 1 {
				log.Fatal("import clone golang error")
			} else {
				generated := fmt.Sprintf("%v %v;\n  %v.value=%v;", v.typ, name, name, v.value)
				if v.typ == "Int" {
					generated += fmt.Sprintf("\n  %v.Float_conv = &Int_to_float;\n  %v.Max_int = &Max_int;\n  %v.Int_to_string = &Int_to_string;\n  %v.Repeat_int = &Repeat_int;\n  %v.Chet = &Chet;\n  %v.Int_length = &Int_length;", name, name, name, name, name, name)
				}
				_, ers := file.WriteString(generated)
				if ers != nil {
					log.Fatal(ers)
				}
			}

		} else if v.typ != "" { // Avoid emitting empty tokens
			if golang == 1 && strings.HasPrefix(v.typ, "net") {
				_, errf := file2.WriteString(fmt.Sprintf("%v\n", v.name))
				if errf != nil {
					log.Fatal(errf)
				}
			} else {
				if v.value == "~" {
					_, errf := file.WriteString(fmt.Sprint(v.name, ";"))
					if errf != nil {
						log.Fatal(errf)
					}

				} else {
					_, errf := file.WriteString(fmt.Sprintf(" %v = %v;\n", v.name, v.value))
					if errf != nil {
						log.Fatal(errf)
					}
				}
			}
		}
	}
}

func main() {
	fmt.Print("name file C: ")
	var name_file string
	_, err := fmt.Scanf("%s", &name_file)
	if err != nil {
		panic("C: false file type")
	}
	name_file += ".c"
	file, er := os.Create(name_file)
	if er != nil {
		panic("no file name")
	}
	fmt.Print("If there is an additional program, enter 1 or 2; otherwise enter 0: ")
	var comanda string
	var comanda2 string
	var znaks int
	fmt.Scan(&znaks)
	if znaks > 0 && znaks < 3 {
		switch znaks {
		case 1:
			fmt.Print("Enter command 1: ")
			reader := bufio.NewReader(os.Stdin)
			comanda, er = reader.ReadString('\n')
			if er != nil {
				log.Fatal(er)
			}
			comanda = strings.TrimSpace(comanda)
		case 2:
			fmt.Print("Enter command 1: ")
			reader := bufio.NewReader(os.Stdin)
			comanda, er = reader.ReadString('\n')
			if er != nil {
				log.Fatal(er)
			}
			comanda = strings.TrimSpace(comanda)
			fmt.Print("Enter command 2: ")
			reader = bufio.NewReader(os.Stdin)
			comanda2, er = reader.ReadString('\n')
			if er != nil {
				log.Fatal(er)
			}
			comanda2 = strings.TrimSpace(comanda2)
		}
	}
	fmt.Print("Specify how long you want the code to run (in seconds). If you want it to shut down immediately after completion, enter “0”. If you want it to run indefinitely, enter “-1”. \n comand: ")
	tim := 0
	_, ersa := fmt.Scanf("%d", &tim)
	if ersa != nil {
		log.Fatal(ersa)
	}
	var tokens []Token
	tokens = Token_length(tokens)

	var file2 *os.File
	var name_file2 string
	if golang > 0 {
		fmt.Print("name file Go: ")
		_, err = fmt.Scanf("%s", &name_file2)
		if err != nil {
			panic("Go: false file type")
		}
		name_file2 += ".go"
		file2, er = os.Create(name_file2)
		if er != nil {
			log.Fatal(er)
		}
		defer file2.Close()
	} else {
		file2, er = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if er != nil {
			log.Fatal(er)
		}
		defer file2.Close()
	}

	sozdanie(tokens, file, file2)

	file.Close()

	fmt.Println("C: Compiling...")
	compileCmd := exec.Command("gcc", name_file, "-o", "my_file", "-lm")
	output, err := compileCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Compilation error:\n%s\n", string(output))
		return
	}

	if golang == 1 {
		fmt.Println("Go: Compiling...")
		modCmd := exec.Command("go", "mod", "init", "eye_module")
		modCmd.Run()

		compileCmd2 := exec.Command("go", "build", "-o", "my_file2", name_file2)
		output2, err2 := compileCmd2.CombinedOutput()
		if err2 != nil {
			fmt.Printf("Compilation error:\n%s\n", string(output2))
			return
		}
	}

	// UDS is required for all net launch modes: Init() connects to it when Go starts.
	var udsListener net.Listener
	if golang == 1 {
		const sockPath = "/tmp/eye.sock"
		if removeErr := os.Remove(sockPath); removeErr != nil && !os.IsNotExist(removeErr) {
			log.Fatal(removeErr)
		}
		var listenErr error
		udsListener, listenErr = net.Listen("unix", sockPath)
		if listenErr != nil {
			log.Fatal(listenErr)
		}
		defer udsListener.Close()
		fmt.Println("UDS server listening on:", sockPath)
	}

	canal := make(chan any)
	canal2 := make(chan any)
	go func() {
		var cmd1 *exec.Cmd
		var cmd2 *exec.Cmd
		var conn net.Conn
		canal_one := make(chan string)
		canal_two := make(chan string)
		if golang == 1 {
			var acceptErr error
			conn, acceptErr = udsListener.Accept()
			if acceptErr != nil {
				log.Fatal(acceptErr)
			}
			defer conn.Close()
			go golang_wirte(conn, canal_two)
			go golang_read(conn, canal_one)
		}
		if tim > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tim)*time.Second)
			defer cancel()
			for {
				log.Println("C: start")
				cmd := exec.CommandContext(ctx, "./my_file")
				stdount, ert := cmd.StdoutPipe()
				if ert != nil {
					log.Fatal(ert)
				}
				cmd.Stderr = os.Stderr
				cStdin, e := cmd.StdinPipe()
				if e != nil {
					log.Fatal(e)
				}
				if golang == 1 {
					go C_read(stdount, canal_two)
					go C_write(cStdin, canal_one)
				} else {
					stdout(stdount)
				}
				switch znaks {
				case 1:
					parts := strings.Fields(comanda)
					cmd_1 := exec.Command(parts[0], parts[1:]...)
					cmd1 = cmd_1
					ers := cmd_1.Start()
					if ers != nil {
						log.Fatal(ers)
					}
				case 2:
					parts := strings.Fields(comanda)
					cmd_1 := exec.Command(parts[0], parts[1:]...)
					cmd1 = cmd_1
					ers := cmd_1.Start()
					if ers != nil {
						log.Fatal(ers)
					}
					parts = strings.Fields(comanda2)
					cmd_2 := exec.Command(parts[0], parts[1:]...)
					cmd2 = cmd_2
					ers = cmd_2.Start()
					if ers != nil {
						log.Fatal(ers)
					}
				}
				ersaf := cmd.Run()
				if cmd1 != nil && cmd1.Process != nil {
					cmd1.Process.Kill()
					cmd1.Process.Release()
				}
				if cmd2 != nil && cmd2.Process != nil {
					cmd2.Process.Kill()
					cmd2.Process.Release()
				}
				if ctx.Err() != nil {
					break
				}
				if ersaf != nil {
					log.Println("C: ", ersaf)
				}
				time.Sleep(4 * time.Second)
			}
		} else if tim == -1 {
			for {
				log.Println("C: start")
				cmd := exec.Command("./my_file")
				cmd.Stderr = os.Stderr
				if golang == 1 {
					stdOutPipe, pipeErr := cmd.StdoutPipe()
					if pipeErr != nil {
						log.Fatal(pipeErr)
					}
					stdinPipe, pipeErr := cmd.StdinPipe()
					if pipeErr != nil {
						log.Fatal(pipeErr)
					}
					go C_read(stdOutPipe, canal_two)
					go C_write(stdinPipe, canal_one)
				} else {
					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
				}
				switch znaks {
				case 1:
					parts := strings.Fields(comanda)
					cmd_1 := exec.Command(parts[0], parts[1:]...)
					cmd1 = cmd_1
					ers := cmd_1.Start()
					if ers != nil {
						log.Fatal(ers)
					}
				case 2:
					parts := strings.Fields(comanda)
					cmd_1 := exec.Command(parts[0], parts[1:]...)
					cmd1 = cmd_1
					ers := cmd_1.Start()
					if ers != nil {
						log.Fatal(ers)
					}
					parts = strings.Fields(comanda2)
					cmd_2 := exec.Command(parts[0], parts[1:]...)
					cmd2 = cmd_2
					ers = cmd_2.Start()
					if ers != nil {
						log.Fatal(ers)
					}
				}
				errror := cmd.Run()
				if cmd1 != nil && cmd1.Process != nil {
					cmd1.Process.Kill()
					cmd1.Process.Release()
				}
				if cmd2 != nil && cmd2.Process != nil {
					cmd2.Process.Kill()
					cmd2.Process.Release()
				}
				if errror != nil {
					log.Println("C: ", errror)
				}
				time.Sleep(5 * time.Second)
			}
		} else {
			log.Println("C: start")
			cmd := exec.Command("./my_file")
			cmd.Stderr = os.Stderr
			if golang == 1 {
				stdOutPipe, pipeErr := cmd.StdoutPipe()
				if pipeErr != nil {
					log.Fatal(pipeErr)
				}
				stdinPipe, pipeErr := cmd.StdinPipe()
				if pipeErr != nil {
					log.Fatal(pipeErr)
				}
				go C_read(stdOutPipe, canal_two)
				go C_write(stdinPipe, canal_one)
			} else {
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
			}
			switch znaks {
			case 1:
				parts := strings.Fields(comanda)
				cmd_1 := exec.Command(parts[0], parts[1:]...)
				cmd1 = cmd_1
				ers := cmd_1.Start()
				if ers != nil {
					log.Fatal(ers)
				}
			case 2:
				parts := strings.Fields(comanda)
				cmd_1 := exec.Command(parts[0], parts[1:]...)
				cmd1 = cmd_1
				ers := cmd_1.Start()
				if ers != nil {
					log.Fatal(ers)
				}
				parts = strings.Fields(comanda2)
				cmd_2 := exec.Command(parts[0], parts[1:]...)
				cmd2 = cmd_2
				ers = cmd_2.Start()
				if ers != nil {
					log.Fatal(ers)
				}
			}
			ers := cmd.Run()
			if cmd1 != nil && cmd1.Process != nil {
				cmd1.Process.Kill()
				cmd1.Process.Release()
			}
			if cmd2 != nil && cmd2.Process != nil {
				cmd2.Process.Kill()
				cmd2.Process.Release()
			}
			if ers != nil {
				fmt.Println("C: ", ers)
			}
		}
		canal <- tim
	}()

	if golang == 1 {
		go func() {
			if tim > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(tim)*time.Second)
				defer cancel()

				for {
					log.Println("Go: ", "start")
					cmd := exec.CommandContext(ctx, "./my_file2")
					cmd.Stderr = os.Stderr
					ersaf := cmd.Run()
					if ctx.Err() != nil {
						break
					}
					if ersaf != nil {
						log.Println("Go: ", ersaf)
					}
					time.Sleep(4 * time.Second)
				}
			} else if tim == -1 {
				for {
					log.Println("Go: ", "start")
					cmd := exec.Command("./my_file2")
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					errror := cmd.Run()
					if errror != nil {
						log.Println("Go: ", errror)
					}
					time.Sleep(5 * time.Second)
				}
			} else {
				log.Println("Go: ", "start")
				cmd := exec.Command("./my_file2")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				ers := cmd.Run()
				if ers != nil {
					fmt.Println("Go: ", ers)
				}
			}
			canal2 <- 1
		}()
	}

	<-canal
	if golang > 0 {
		<-canal2
	}
}
func C_read(stdout io.Reader, canal chan string) {
	reader := bufio.NewReader(stdout)
	for {

		line, er := reader.ReadString('\n')
		if er != nil {
			return
		}
		if !strings.HasPrefix(line, "$") {
			fmt.Println(line)
			continue
		}
		canal <- line[1:]
	}

}
func C_write(stdin io.Writer, canal chan string) {
	for value := range canal {
		value = strings.TrimRight(value, "\r\n")
		if _, er := fmt.Fprintln(stdin, value); er != nil {
			log.Println(er)
			return
		}
	}
}
func golang_read(conn net.Conn, canal chan string) {
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(line)
		canal <- line
	}
}
func golang_wirte(conn net.Conn, canal chan string) {
	for value := range canal {
		value = strings.TrimRight(value, "\r\n")
		if _, er := fmt.Fprintln(conn, value); er != nil {
			log.Println(er)
			return
		}
	}
}
func stdout(stdout io.Reader) {
	reader := bufio.NewReader(stdout)
	for {

		line, er := reader.ReadString('\n')
		if er != nil {
			log.Println(er)
			return
		}
		fmt.Println(line)
	}
}
