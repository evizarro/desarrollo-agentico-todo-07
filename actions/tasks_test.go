package actions

import (
	"net/url"
	"strings"

	"todo/models"

	"github.com/gofrs/uuid"
)

// day devuelve el día de hoy desplazado n días, con el formato del campo de fecha.
func day(n int) string {
	return models.Today().AddDate(0, 0, n).Format(models.DueDateLayout)
}

// dayDisplay devuelve el día de hoy desplazado n días, como lo muestra la lista (DD/MM/AAAA).
func dayDisplay(n int) string {
	return models.Today().AddDate(0, 0, n).Format("02/01/2006")
}

// savedDay devuelve el día guardado de una tarea, con el formato del campo de fecha.
func savedDay(task *models.Task) string {
	return task.DueDate.Format(models.DueDateLayout)
}

func (as *ActionSuite) Test_TasksIndex_EmptyState() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "No hay tareas todavía")
}

func (as *ActionSuite) Test_TasksIndex_ListsTasks() {
	as.LoadFixture("tres tareas")

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Comprar material de oficina")
	as.Contains(body, "Llamar al gestor")
	as.Contains(body, "Publicar la oferta de empleo")
	as.Contains(body, "Completada")
	as.Contains(body, "01/03/2099")
	as.Contains(body, "02/03/2099")
	as.Contains(body, "03/03/2099")
	as.NotContains(body, "Vencida")
}

func (as *ActionSuite) Test_TasksIndex_NewTaskFormHasDueDate() {
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, `type="date"`)
	as.Contains(body, `name="due_date"`)
	as.Contains(body, `for="task-due-date"`)
	as.Contains(body, "Fecha límite")
}

func (as *ActionSuite) Test_TasksCreate_Valid() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"  Nueva tarea de prueba  "}, "due_date": {day(1)}})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Nueva tarea de prueba", task.Title)
	as.False(task.Completed)
	as.Equal(day(1), savedDay(task))
}

func (as *ActionSuite) Test_TasksCreate_WithDueDate() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Entregar el informe"}, "due_date": {day(1)}})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(1, count)

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal("Entregar el informe", task.Title)
	as.False(task.Completed)
	as.Equal(day(1), savedDay(task))

	res = as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Tarea creada.")
	as.Contains(res.Body.String(), dayDisplay(1))
}

func (as *ActionSuite) Test_TasksCreate_EmptyTitle_Invalid() {
	res := as.HTML("/tasks").Post(url.Values{"title": {""}, "due_date": {day(1)}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_OnlySpaces_Invalid() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"     "}, "due_date": {day(1)}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_DueDateRequired() {
	empty := url.Values{"title": {"Sin fecha"}, "due_date": {""}}
	absent := url.Values{"title": {"Sin fecha"}}
	for _, form := range []url.Values{empty, absent} {
		res := as.HTML("/tasks").Post(form)
		as.Equal(422, res.Code)

		body := res.Body.String()
		as.Contains(body, "La fecha límite es obligatoria.")
		as.Contains(body, `id="due-date-error"`)
		as.Contains(body, `value="Sin fecha"`)
	}

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_DueDateInvalid() {
	for _, raw := range []string{"mañana", "2027-02-30", "30/12/2027"} {
		res := as.HTML("/tasks").Post(url.Values{"title": {"Fecha mala"}, "due_date": {raw}})
		as.Equal(422, res.Code)
		as.Contains(res.Body.String(), "La fecha límite no es una fecha válida.")
	}

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_PastDueDate_Invalid() {
	for _, date := range []string{day(-1), day(-365)} {
		res := as.HTML("/tasks").Post(url.Values{"title": {"Fecha pasada"}, "due_date": {date}})
		as.Equal(422, res.Code)

		body := res.Body.String()
		as.Contains(body, "La fecha límite no puede ser anterior a hoy.")
		as.Contains(body, `value="Fecha pasada"`)
		as.Contains(body, `value="`+date+`"`)
	}

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksCreate_TodayDueDate_Valid() {
	res := as.HTML("/tasks").Post(url.Values{"title": {"Vence hoy"}, "due_date": {day(0)}})
	as.Equal(303, res.Code)

	task := &models.Task{}
	as.NoError(as.DB.First(task))
	as.Equal(day(0), savedDay(task))
}

func (as *ActionSuite) Test_TasksToggle_CompletesAndReopens() {
	task := &models.Task{Title: "Pendiente de completar", DueDate: models.Today().AddDate(0, 0, 1)}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)

	res = as.HTML("/tasks/%s/toggle", task.ID).Put(url.Values{})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksToggle_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s/toggle", id).Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksToggle_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/toggle").Put(url.Values{})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksDestroy() {
	task := &models.Task{Title: "Tarea que se elimina", DueDate: models.Today().AddDate(0, 0, 1)}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/tasks/%s", task.ID).Delete()
	as.Equal(303, res.Code)

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(0, count)
}

func (as *ActionSuite) Test_TasksDestroy_NotFound() {
	id := uuid.Must(uuid.NewV4())
	res := as.HTML("/tasks/%s", id).Delete()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) taskForEditing(completed bool) *models.Task {
	task := &models.Task{Title: "Título original", Completed: completed, DueDate: models.Today().AddDate(0, 0, 1)}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	as.NoError(as.DB.Reload(task))
	return task
}

// pastDueTask inserta, sin validar, una tarea cuya fecha límite fue ayer.
func (as *ActionSuite) pastDueTask() *models.Task {
	task := &models.Task{Title: "Título original", DueDate: models.Today().AddDate(0, 0, -1)}
	as.NoError(as.DB.Create(task))
	as.NoError(as.DB.Reload(task))
	return task
}

func (as *ActionSuite) Test_TasksIndex_EditLink() {
	task := as.taskForEditing(false)
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "/tasks/"+task.ID.String()+"/edit")
	as.Contains(res.Body.String(), "Editar tarea: Título original")
}

func (as *ActionSuite) Test_TasksIndex_DueDatesAndOverdue() {
	// Se crean en un orden distinto del de sus fechas: la lista sigue el orden de creación.
	cases := []*models.Task{
		{Title: "Pendiente de hoy", DueDate: models.Today()},
		{Title: "Pendiente de ayer", DueDate: models.Today().AddDate(0, 0, -1)},
		{Title: "Pendiente de mañana", DueDate: models.Today().AddDate(0, 0, 1)},
		{Title: "Completada de ayer", DueDate: models.Today().AddDate(0, 0, -1), Completed: true},
	}
	for _, task := range cases {
		as.NoError(as.DB.Create(task))
	}

	res := as.HTML("/").Get()
	as.Equal(200, res.Code)

	rows := strings.Split(res.Body.String(), `<li class="task`)[1:]
	as.Len(rows, 4)

	expected := []struct {
		title, date        string
		overdue, pastDueAt bool
	}{
		{"Pendiente de hoy", dayDisplay(0), false, false},
		{"Pendiente de ayer", dayDisplay(-1), true, true},
		{"Pendiente de mañana", dayDisplay(1), false, false},
		{"Completada de ayer", dayDisplay(-1), false, true},
	}
	for i, want := range expected {
		as.Contains(rows[i], want.title)
		as.Contains(rows[i], want.date)
		as.Equal(want.overdue, strings.Contains(rows[i], "Vencida"), want.title)
		as.Equal(want.pastDueAt, strings.Contains(rows[i], `data-past-due="true"`), want.title)
	}
}

func (as *ActionSuite) Test_TasksEdit_PrefillsTitle() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), `value="Título original"`)
	as.Contains(res.Body.String(), `value="PUT"`)
	as.Contains(res.Body.String(), "Guardar cambios")
	as.Contains(res.Body.String(), "Cancelar")
	as.NoError(as.DB.Reload(task))
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksEdit_PrefillsDueDate() {
	for _, completed := range []bool{false, true} {
		task := as.taskForEditing(completed)
		res := as.HTML("/tasks/%s/edit", task.ID).Get()
		as.Equal(200, res.Code)

		body := res.Body.String()
		as.Contains(body, `name="due_date"`)
		as.Contains(body, `value="`+savedDay(task)+`"`)
	}
}

func (as *ActionSuite) Test_TasksEdit_NotFound() {
	res := as.HTML("/tasks/%s/edit", uuid.Must(uuid.NewV4())).Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid/edit").Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksEdit_EscapesTitle() {
	task := &models.Task{Title: `<img src=x onerror="alert(1)">`, DueDate: models.Today().AddDate(0, 0, 1)}
	verrs, err := as.DB.ValidateAndCreate(task)
	as.NoError(err)
	as.False(verrs.HasAny())
	res := as.HTML("/tasks/%s/edit", task.ID).Get()
	as.Equal(200, res.Code)
	as.NotContains(res.Body.String(), `<img src=x`)
	as.Contains(res.Body.String(), "&lt;img")
}

func (as *ActionSuite) Test_TasksUpdate_PersistsTitleOnly() {
	task := as.taskForEditing(true)
	id, createdAt, updatedAt, dueDate := task.ID, task.CreatedAt, task.UpdatedAt, savedDay(task)
	res := as.HTML("/tasks/%s", id).Put(url.Values{
		"title": {"  Título actualizado  "}, "due_date": {dueDate}, "completed": {"false"},
		"id": {uuid.Must(uuid.NewV4()).String()}, "created_at": {"2000-01-01T00:00:00Z"},
	})
	as.Equal(303, res.Code)
	as.Equal("/", res.Header().Get("Location"))
	as.NoError(as.DB.Reload(task))
	as.Equal("Título actualizado", task.Title)
	as.Equal(id, task.ID)
	as.True(task.Completed)
	as.True(task.CreatedAt.Equal(createdAt))
	as.False(task.UpdatedAt.Before(updatedAt))
	as.Equal(dueDate, savedDay(task))
	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(1, count)
}

func (as *ActionSuite) Test_TasksUpdate_PendingTaskRemainsPending() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{
		"title": {"Nuevo título"}, "due_date": {savedDay(task)}, "completed": {"true"},
	})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal("Nuevo título", task.Title)
	as.False(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyTitle() {
	task := as.taskForEditing(false)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {""}, "due_date": {savedDay(task)}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.Contains(res.Body.String(), `aria-describedby="title-error"`)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_OnlySpaces() {
	task := as.taskForEditing(true)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"   "}, "due_date": {savedDay(task)}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título es obligatorio.")
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
	as.True(task.Completed)
}

func (as *ActionSuite) Test_TasksUpdate_TitleTooLong() {
	task := as.taskForEditing(false)
	title := strings.Repeat("á", models.TitleMaxLength+1)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}, "due_date": {savedDay(task)}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "El título no puede superar los 200 caracteres.")
	as.Contains(res.Body.String(), title)
	as.NoError(as.DB.Reload(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_TitleAtLimit() {
	task := as.taskForEditing(false)
	dueDate := savedDay(task)
	title := strings.Repeat("á", models.TitleMaxLength)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {title}, "due_date": {dueDate}})
	as.Equal(303, res.Code)
	as.NoError(as.DB.Reload(task))
	as.Equal(title, task.Title)
	as.Equal(dueDate, savedDay(task))
}

func (as *ActionSuite) Test_TasksUpdate_ChangesDueDate() {
	for _, completed := range []bool{false, true} {
		task := as.taskForEditing(completed)
		createdAt := task.CreatedAt
		res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {task.Title}, "due_date": {day(7)}})
		as.Equal(303, res.Code)

		as.NoError(as.DB.Reload(task))
		as.Equal(day(7), savedDay(task))
		as.Equal("Título original", task.Title)
		as.Equal(completed, task.Completed)
		as.True(task.CreatedAt.Equal(createdAt))
	}

	count, err := as.DB.Count(&models.Task{})
	as.NoError(err)
	as.Equal(2, count)
}

func (as *ActionSuite) Test_TasksUpdate_EmptyDueDate() {
	task := as.taskForEditing(false)
	dueDate := savedDay(task)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título nuevo"}, "due_date": {""}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "La fecha límite es obligatoria.")
	as.NoError(as.DB.Reload(task))
	as.Equal(dueDate, savedDay(task))
	as.Equal("Título original", task.Title)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidDueDate() {
	task := as.taskForEditing(false)
	dueDate := savedDay(task)
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título nuevo"}, "due_date": {"2027-02-30"}})
	as.Equal(422, res.Code)
	as.Contains(res.Body.String(), "La fecha límite no es una fecha válida.")
	as.NoError(as.DB.Reload(task))
	as.Equal(dueDate, savedDay(task))
}

func (as *ActionSuite) Test_TasksUpdate_KeepsPastDueDate() {
	task := as.pastDueTask()
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {"Título nuevo"}, "due_date": {day(-1)}})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.Equal("Título nuevo", task.Title)
	as.Equal(day(-1), savedDay(task))

	res = as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Vencida")
}

func (as *ActionSuite) Test_TasksUpdate_OtherPastDueDate() {
	task := as.pastDueTask()
	res := as.HTML("/tasks/%s", task.ID).Put(url.Values{"title": {task.Title}, "due_date": {day(-7)}})
	as.Equal(303, res.Code)

	as.NoError(as.DB.Reload(task))
	as.Equal(day(-7), savedDay(task))
}

func (as *ActionSuite) Test_TasksUpdate_NotFound() {
	res := as.HTML("/tasks/%s", uuid.Must(uuid.NewV4())).Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_TasksUpdate_InvalidID() {
	res := as.HTML("/tasks/no-es-un-uuid").Put(url.Values{"title": {"Título nuevo"}})
	as.Equal(404, res.Code)
}
