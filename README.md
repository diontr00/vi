# Vi Router

"This project is inspired by **gorilla/mux** and aims to create a high-performance HTTP router that offers powerful and convenient routing capabilities that support regex. But somehow that project is quite slow, so i plan out to create something similar but better in perfomance by reviewing the implementation of **julienschmidt/httprouter**. I hope i can receive some feedback cause this is my first Open Source project.

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
