# Añadir fechas límite a las tareas

**Issue:** #1 · **Tipo:** feature

## Resumen

Cada tarea lleva ahora una fecha límite obligatoria y editable. Se fija al crear la tarea y se cambia desde la pantalla de edición; la lista la muestra y marca como «Vencida» las tareas pendientes cuya fecha ya pasó. La petición era «Añadir fechas límite a las tareas que sean editables»; las dudas se resolvieron con el humano (ver [decisiones](../decisions/feature-1-anadir-fechas-limite-a-las-tareas.md)).

## Qué se construyó

- Campo `due_date` en `tasks`: un día, sin hora, obligatorio.
- Formulario «Nueva tarea» con campo de fecha, y pantalla de edición con la fecha además del título.
- Validación: la fecha es obligatoria y debe ser un día válido; al crear no puede ser anterior a hoy; al editar se acepta cualquier fecha válida, también pasada.
- Lista en orden de creación (sin cambio), con la fecha junto al título y la etiqueta «Vencida» en las tareas pendientes con fecha pasada. Una tarea completada no se marca como vencida.
- Marcar o reabrir una tarea sin recargar actualiza también la etiqueta «Vencida».
- Arreglo del toggle con JavaScript (ver Notas, excepción de contención).

## Implementación técnica

| Fichero | Papel |
|---|---|
| `migrations/20261021000000_add_due_date_to_tasks.{up,down}.fizz` | Añade y quita `due_date`. Las tareas existentes reciben `CURRENT_DATE`; la columna es `NOT NULL`. |
| `models/task.go` | `DueDate`, `SetDueDate`, `Validate` (obligatoria, formato `2006-01-02`), `ValidateCreate` (no anterior a hoy), `PastDue`, `Overdue`, `DueDateValue`. |
| `actions/tasks.go` | `taskInput` compartido por crear y actualizar (título y `due_date`); ambos llaman a `SetDueDate`. |
| `templates/tasks/index.plush.html`, `edit.plush.html` | Campo `<input type="date">`, errores por campo (`title` y `due_date` por separado), fecha y «Vencida» en la lista. |
| `public/assets/app.js`, `app.css` | Alterna «Vencida» al completar/reabrir sin recargar; estilos de fecha y etiqueta. |

Cambios que importan:
- La regla de fecha y la de «vencida» viven en el modelo. El JS solo lee `data-past-due`, calculado por el servidor.
- `Today()` usa el reloj de la aplicación a medianoche UTC; `PastDue` compara solo el día.
- `ValidateCreate` es el hook de pop que solo corre al crear; por eso editar conserva una fecha ya pasada.
- Un texto de fecha inválido deja `DueDate` a cero y se distingue de «vacío» por `dueDateInput` (campo no persistido).
- Documentación viva actualizada en el propio cambio: `docs/knowledge/overview.md`, `docs/knowledge/features/tareas.md`, `docs/procedures/development.md`.

## Cómo se usa

1. En «Nueva tarea», escribir el título y elegir la fecha; si falta o es pasada, se rechaza con un mensaje junto al campo.
2. En la lista, cada tarea muestra su fecha (`dd/mm/aaaa`); si está pendiente y vencida, lleva «Vencida».
3. En editar, se cambia título y fecha. Se puede dejar una fecha ya pasada, pero no vaciarla.

## Configuración

Requiere aplicar la migración `20261021000000_add_due_date_to_tasks` (`soda migrate`). Las tareas ya guardadas reciben `CURRENT_DATE` como fecha límite: las pendientes antiguas aparecerán «Vencida» a partir del día siguiente.

## Cómo se probó

Validación ejecutada por el humano en una sesión con aprobación real, sobre el commit `6665df2`: [acta de validación](https://github.com/evizarro/desarrollo-agentico-todo-07/issues/1#issuecomment-6321907427).

| Afirmación | Resultado |
|---|---|
| Crear con fecha; editar la fecha sin tocar título ni estado | Cumple (`303`) |
| Fecha vacía, ausente o mal formada (`mañana`, `2027-02-30`, `30/12/2027`) al crear y editar | Cumple (`422` con mensaje, nunca `500`) |
| Migración sobre 4 tareas existentes: se conservan y reciben `CURRENT_DATE`; la base rechaza inserciones sin fecha | Cumple |
| Fecha pasada rechazada al crear (ayer, hace un año); hoy aceptada | Cumple |
| Editar conserva una fecha vencida y sigue «Vencida» | Cumple |
| Lista: fecha, «Vencida» solo en pendientes vencidas, orden de creación | Cumple |
| Completar/reabrir sin recargar (navegador) | Cumple tras el arreglo: 8/8 en verde |
| Sin regresiones: `todo/actions`, `todo/models` y Playwright 6/6 | Cumple |
| Juicios: textos, trazabilidad criterio-test, documentación viva, estructura | Cumplen sin reservas |

Salvedades del acta, tal como constan:
- **`gofmt` falla**, con exactamente los mismos 15 ficheros que marca en `main` sin tocar este cambio; ninguno es de este cambio. Formalmente el criterio de estructura falla si falla uno de los cuatro gates primeros. El humano lo aceptó como no bloqueante; se arreglará como tarea de mantenimiento aparte.
- **`todo/adws` falla con unos 25 tests en rojo**, ninguno relacionado con esta issue: fallos de Windows nativo (`gh`, `bin/adw`, binarios de prueba sin extensión que `exec.LookPath` no resuelve) y de configuración de las skills, anteriores al cambio.
- **Contención (A15): no cumple del todo, 1 blocker.** El arreglo del toggle corrige un comportamiento existente que ningún criterio pide; ver Notas.

`vet`, `migrate` y `build` en verde.

Para repetir:

```
GO_ENV=test go test -p 1 ./actions/... ./models/...
npx playwright test
```

Hay que cargar antes el entorno del worktree (ver `docs/procedures/development.md`).

## Notas

- **Decisiones del humano:** [feature-1-anadir-fechas-limite-a-las-tareas.md](../decisions/feature-1-anadir-fechas-limite-a-las-tareas.md).
- **Decisión sobre el acta** ([comentario](https://github.com/evizarro/desarrollo-agentico-todo-07/issues/1#issuecomment-6322170351)): se acepta `CURRENT_DATE` como fecha por defecto de las tareas existentes; se acepta la excepción de contención; el `gofmt` preexistente no bloquea. La validación queda aceptada.
- **Excepción de contención:** `public/assets/app.js` hacía `fetch(form.action)` sin método, es decir `GET`, contra una ruta que solo existe como `PUT`: un bug preexistente a esta issue, por el que el toggle con JavaScript nunca persistía. Se arregló (`390837c`) reenviando el formulario con `FormData`, a petición expresa del humano, porque sin ello no se podía verificar el criterio de «sin recargar». Un segundo arreglo (`6665df2`) hace que el test de navegador espere la respuesta de `/toggle` antes de recargar: había una carrera, en rojo 1 de 2 antes y 8 de 8 en verde después.
- **Límites:** la fecha es un día sin hora y se compara con el reloj de la aplicación en UTC. Editar a otra fecha pasada distinta de la anterior se acepta, como decidió el plan. La lista no se ordena por fecha.
