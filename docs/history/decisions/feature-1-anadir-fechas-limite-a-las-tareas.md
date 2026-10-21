# Decisiones de Añadir fechas límite a las tareas

**Issue:** #1 · **Documento de cierre:** [feature-1-anadir-fechas-limite-a-las-tareas.md](../changes/feature-1-anadir-fechas-limite-a-las-tareas.md)

Las cuatro preguntas se plantearon en un [comentario de aclaración](https://github.com/evizarro/desarrollo-agentico-todo-07/issues/1#issuecomment-6320286075). El humano respondió en un solo comentario: «1 A, 2 B, 3 C, 4 B».

## ¿Cuándo se puede poner o cambiar la fecha límite?

- **Contexto:** la petición dice que las fechas sean editables, pero no si se fijan al crear.
- **Opciones:** A. Al crear y al editar. B. Solo al editar; las tareas nuevas nacen sin fecha.
- **Elección:** «1 A» · [comentario](https://github.com/evizarro/desarrollo-agentico-todo-07/issues/1#issuecomment-6320367991)
- **Consecuencia:** el formulario «Nueva tarea» tiene un campo de fecha y la pantalla de edición también.

## ¿Es obligatoria la fecha límite?

- **Contexto:** determina si hay tareas sin fecha y qué pasa con las ya guardadas.
- **Opciones:** A. No; se puede no tener fecha y quitarla. B. Sí; las tareas guardadas reciben una fecha por defecto, a decidir.
- **Elección:** «2 B» · [comentario](https://github.com/evizarro/desarrollo-agentico-todo-07/issues/1#issuecomment-6320367991)
- **Consecuencia:** una fecha vacía se rechaza al crear y al editar, y la columna es `NOT NULL`. La migración da `CURRENT_DATE` a las tareas existentes (aceptado por el humano en la [decisión sobre el acta](https://github.com/evizarro/desarrollo-agentico-todo-07/issues/1#issuecomment-6322170351)).

## ¿Se admiten fechas pasadas?

- **Contexto:** una fecha pasada al crear suele ser un error, pero una tarea vencida debe poder editarse.
- **Opciones:** A. Sí, sin restricción. B. No, ni al crear ni al editar. C. Solo se rechazan al crear; al editar se conserva la que tenía aunque haya pasado.
- **Elección:** «3 C» · [comentario](https://github.com/evizarro/desarrollo-agentico-todo-07/issues/1#issuecomment-6320367991)
- **Consecuencia:** crear con fecha anterior a hoy da `422`. Editar acepta cualquier fecha válida, incluida otra fecha pasada.

## ¿Cómo se muestra en la lista?

- **Contexto:** define lo que ve el usuario y el orden de la lista.
- **Opciones:** A. Solo la fecha junto al título. B. La fecha y la etiqueta «Vencida» en las pendientes vencidas. C. Lo de B y orden por fecha límite.
- **Elección:** «4 B» · [comentario](https://github.com/evizarro/desarrollo-agentico-todo-07/issues/1#issuecomment-6320367991)
- **Consecuencia:** la lista muestra la fecha y «Vencida» en las pendientes con fecha pasada, y sigue en orden de creación. Las completadas no se marcan como vencidas.
