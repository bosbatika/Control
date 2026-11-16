# Control

[![Telegram](https://img.shields.io/badge/Telegram-Join_Chat-blue?logo=telegram&style=for-the-badge)](https://t.me/control_eye)
![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Status](https://img.shields.io/badge/status-early_alpha-orange)
![License](https://img.shields.io/badge/license-Apache_2.0-blue)

**Control** is a hybrid programming language designed to combine the performance of C with the networking capabilities of Go. 

Source files written in Control have the **`.eye`** extension. The compiler is currently distributed as a single-file monolithic Go script for easy setup.

## 🚀 Key Features

*   **Dual-Backend:** Transpiles `.eye` files into C. If `import net` is used, it also generates a Go web server.
*   **Fault Tolerance (Supervisor Pattern):** The runtime launches the compiled C and Go executables as independent processes and monitors them. If one crashes, it automatically restarts it.
*   **IPC Bridge:** C and Go processes communicate via a Unix Domain Socket (UDS). The C process writes to `stdout` with a `$` prefix, which the Go wrapper intercepts and sends via UDS.
*   **No Garbage Collector:** Memory management is manual (via C's `malloc`/`free`), making it suitable for low-level hardware control.
*   **Method Binding:** Control types (like `Int`, `String`) are transpiled into C structs with bound function pointers.

## 📦 Code Example (`.eye` file)

```eye
import stop_import ;

Int chislo.Start = 12 ;
@ printf("Number: %d\n", chislo.value) ;
@ printf("Is even? %d\n", chislo.Chet(chislo)) ;

end
