# Control

[![Telegram](https://img.shields.io/badge/Telegram-Join_Chat-blue?logo=telegram&style=for-the-badge)](https://t.me/control_eye)
![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Status](https://img.shields.io/badge/status-early_alpha-orange)
![License](https://img.shields.io/badge/license-Apache_2.0-blue)

**Control** is a compiler and runtime manager for the **EYE** programming language.

## Terminology
To avoid confusion:
*   **Control**: The compiler and runtime environment (written in Go).
*   **EYE**: The programming language itself.
*   **.eye**: The file extension for EYE source code.

## What is EYE?
EYE is a hybrid programming language designed to combine the performance of C with the networking capabilities of Go. Control compiles `.eye` files into C code, and optionally into Go code if the `net` module is used.

## 🚀 Key Features

*   **Dual-Backend:** Transpiles `.eye` files into C. If `import net` is used, it also generates a Go web server.
*   **Fault Tolerance (Supervisor Pattern):** The Control runtime launches the compiled C and Go executables as independent processes and monitors them. If one crashes, Control automatically restarts it.
*   **IPC Bridge:** C and Go processes communicate via a Unix Domain Socket (UDS). The C process writes to `stdout` with a `$` prefix, which the Go wrapper intercepts and sends via UDS.
*   **No Garbage Collector:** Memory management is manual (via C's `malloc`/`free`), making it suitable for low-level hardware control.
*   **Method Binding:** EYE types (like `Int`, `String`) are transpiled into C structs with bound function pointers.

## 📦 EYE Code Example (`.eye` file)

```eye
import stop_import ;

Int chislo.Start = 12 ;
@ printf("Number: %d\n", chislo.value) ;
@ printf("Is even? %d\n", chislo.Chet(chislo)) ;

end
