## 🏗️ Architecture

Control is built on a modular, multi-language architecture designed to leverage the strengths of both C and Go. It uses a dual-backend approach to separate high-performance hardware control from high-level networking.

### 1. Compilation Pipeline
The compiler reads `.eye` source files and converts them into an Abstract Syntax Tree (AST). Then, it routes the AST to specific code generators:
*   **C Generator (Core Backend):** Translates core logic, types, and control structures into standard C code. It maps EYE types (e.g., `Int`, `String`) directly to C structs with specific method pointers.
*   **Go Generator (Network Backend):** Triggered only when `import net` is used. It generates a Go wrapper to handle HTTP requests, async networking, and WebSockets.

### 2. Runtime & IPC Bridge (The Secret Sauce)
When both C and Go components are present, they run as separate processes to prevent crashes in one from affecting the other. 
*   **Unix Domain Socket (UDS):** Go acts as the network listener (HTTP server). When a request arrives, Go forwards it through a UDS bridge.
*   **stdin/stdout Bridge:** The C process reads incoming data from standard input (`stdin`) and writes responses to standard output (`stdout`). The compiler's runtime wrapper routes these streams through the UDS.
*   **The `$` Marker:** The C part sends responses starting with a `$` prefix (e.g., `$ok` or `$text/plain!Hello`), which the bridge intercepts, strips, and passes back to Go to form a proper HTTP response.

### 3. Type System & Memory
*   **C-Native Types:** Types like `Int` and `Float` are literally C structures (`struct`) with a `value` field.
*   **Method Binding:** When a variable is declared using `.Start`, the compiler automatically assigns the correct function pointers (methods) to its struct. For example, `chislo.Chet()` directly calls a C function bound to the `Int` type.
*   **Manual Memory Control:** Using the `@` prefix, the developer can directly call C standard library functions like `@ malloc()` or `@ free()`, bridging the gap between EYE and low-level memory management.
