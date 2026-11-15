# Control Compiler (for EYE Language)

[![Telegram](https://img.shields.io/badge/Telegram-Join_Chat-blue?logo=telegram&style=for-the-badge)](https://t.me/control_eye)
![Status](https://img.shields.io/badge/status-early_alpha-orange)
![License](https://img.shields.io/badge/license-Apache_2.0-blue)
![Backend](https://img.shields.io/badge/backend-C_&_Go-green)

**Control(version 0.1)** is a hybrid compiler and runtime environment for the **EYE** programming language. It is designed to unify the worlds of low-level hardware control and high-level networking into a single, cohesive ecosystem. 

Source files written in EYE use the **`.eye`** extension. Currently, the Control compiler is distributed as a **single-file monolithic Go script** for easy setup and rapid testing.

## 💡 Motivation & Philosophy

Developers often have to juggle multiple languages: C for embedded systems and performance, Go for backend and networking, and Elixir (or similar) for fault tolerance. **Control** was created to solve this fragmentation. 

The goal is to allow developers to learn and write in *one* language (**EYE**) while utilizing the strengths of others under the hood:
*   **The Power of C:** Direct memory management, maximum performance, and direct hardware access.
*   **The Network of Go:** Built-in, high-performance HTTP servers and async networking.
*   **The Resilience of Elixir:** Inspired by Erlang/Elixir's "let it crash" philosophy, Control features a runtime manager that isolates the C and Go processes and automatically restarts them if they fail. 

## 🚀 Key Features

*   **Dual-Backend (C & Go):** Control transpiles `.eye` files into C code. If `import net` is detected, it also generates a Go web server.
*   **Elixir-Inspired Fault Tolerance:** The runtime manager launches the compiled C and Go executables as independent processes and monitors them. If one crashes, it is automatically restarted.
*   **Built-in IPC Bridge:** C and Go processes communicate via a Unix Domain Socket (UDS). The C process reads `stdin` and writes to `stdout` (using the `$` prefix), which the Go wrapper intercepts and sends via UDS.
*   **Method Binding:** EYE types (like `Int`, `String`) are transpiled into C structs with bound function pointers (e.g., `value.Chet()`), bringing a touch of high-level OOP to bare C.
*   **Minimalist Syntax:** Uses `.Start` for declarations, `^` for type repetition, and `@` to directly call C functions.

## 📦 Code Example (`.eye` file)

```eye
import stop_import ;

Int chislo.Start = 12 ;
@ printf("Number: %d\n", chislo.value) ;
@ printf("Is even? %d\n", chislo.Chet(chislo)) ;

end
