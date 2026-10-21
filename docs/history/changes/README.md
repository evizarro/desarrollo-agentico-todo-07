# Cambios documentados

Cada caso que cierra el workflow `adw` deja aquí su documento de cierre: qué se construyó, cómo se usa, cómo se probó y qué conviene saber para trabajar sobre ello. Lo escribe la skill `document` en la rama del caso, y viaja en el PR con el código.

Este índice es condicional: cada entrada dice cuándo conviene cargar su documento. Se llega a él desde [`docs/README.md`](../../README.md): quien planifica, como las skills `feature`, `bug` y `chore`, lee solo los documentos cuyas condiciones encajan con el encargo.

Un documento de cierre cuenta qué pasó en su caso, y no se reescribe cuando otro caso cambia el mismo código. Lo vigente está en `docs/knowledge/` y `docs/procedures/`.

## Formato de las entradas

```markdown
- [<tipo>-<N>-<slug>.md](<tipo>-<N>-<slug>.md): <qué cambió, en una línea>
  - Cuando se trabaje con <rutas, acciones o plantillas concretas>
  - Cuando se modifique <modelo, campo o regla relacionados>
  - Cuando se investigue <un problema concreto de esa área>
```

Entre tres y cinco condiciones por documento, y las entradas en orden por número de issue. Si un documento ya tiene entrada, se actualizan sus condiciones en lugar de añadir otra.

## Índice

- [feature-1-anadir-fechas-limite-a-las-tareas.md](feature-1-anadir-fechas-limite-a-las-tareas.md): fecha límite obligatoria y editable en las tareas, con etiqueta «Vencida» en la lista
  - Cuando se trabaje con `models/task.go`, `actions/tasks.go` o las plantillas `templates/tasks/index.plush.html` y `edit.plush.html`
  - Cuando se modifique la tabla `tasks`, la columna `due_date` o su validación (obligatoria, no pasada al crear)
  - Cuando se toque `public/assets/app.js` o el toggle de completar/reabrir sin recargar
  - Cuando se investigue por qué una tarea sale «Vencida» o por qué las tareas antiguas tienen la fecha de la migración
