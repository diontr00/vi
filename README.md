# Vi Router

[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![Go Reference](https://pkg.go.dev/badge/github.com/diontr00/vi.svg)](https://pkg.go.dev/github.com/diontr00/vi)
![ci workflow](https://github.com/diontr00/gocolor/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/diontr00/vi/graph/badge.svg?token=bPz6VDXHae)](https://codecov.io/gh/diontr00/vi)

"This project draws inspiration from **gorilla/mux**, and I appreciate the ease of working with the library. However, its performance falls short of my expectations. After conducting research, my goal is to develop a high-performance HTTP router that not only outperforms but also retains the convenient API of mux, enhanced with additional support for regex. I welcome any feedback as this will be my first open source projects."

# Matching Rule

- **Named parameter**
  - **Syntax:** :name
  - **Example:** /student/:name
  - **Explain:** Match name as word

* **Name with regex pattern**
  - **Syntax:** {name:regex-pattern}
  - **Example:** /student/{id:[0-9]+}
  * **Explain:** Match id as number

- **Helper pattern**
  - **:id** : short for **/student/{id:[0-9]+}**
  - **:name** : short for **/{name:[0-9a-zA-Z]+}**
