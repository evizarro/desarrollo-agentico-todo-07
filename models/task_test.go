package models

import (
	"strings"
	"time"
)

// daysFromToday devuelve el día de hoy desplazado n días.
func daysFromToday(n int) time.Time {
	return Today().AddDate(0, 0, n)
}

func (ms *ModelSuite) Test_Task_Create_Valid() {
	task := &Task{Title: "Comprar café", DueDate: daysFromToday(1)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.NotZero(task.ID)
	ms.False(task.Completed)

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(1, count)
}

func (ms *ModelSuite) Test_Task_Title_Required() {
	task := &Task{Title: ""}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_Title_OnlySpaces_Invalid() {
	task := &Task{Title: "   \t  "}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.Contains(verrs.Get("title"), "El título es obligatorio.")
}

func (ms *ModelSuite) Test_Task_Title_MaxLength() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength), DueDate: daysFromToday(1)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())
}

func (ms *ModelSuite) Test_Task_Title_TooLong_Invalid() {
	task := &Task{Title: strings.Repeat("a", TitleMaxLength+1)}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.True(verrs.HasAny())
	ms.NotEmpty(verrs.Get("title"))
}

func (ms *ModelSuite) Test_Task_Toggle_Persists() {
	task := &Task{Title: "Revisar el correo", DueDate: daysFromToday(1)}
	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	task.Completed = true
	ms.NoError(ms.DB.Update(task))

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.True(reloaded.Completed)
}

func (ms *ModelSuite) Test_Task_DueDate_Required() {
	for _, raw := range []string{"", "   "} {
		withoutDate := &Task{Title: "Sin fecha"}
		withoutDate.SetDueDate(raw)

		verrs, err := ms.DB.ValidateAndCreate(withoutDate)
		ms.NoError(err)
		ms.True(verrs.HasAny())
		ms.Contains(verrs.Get("due_date"), "La fecha límite es obligatoria.")
	}

	untouched := &Task{Title: "Sin fecha"}
	verrs, err := ms.DB.ValidateAndCreate(untouched)
	ms.NoError(err)
	ms.Contains(verrs.Get("due_date"), "La fecha límite es obligatoria.")

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_DueDate_Invalid() {
	for _, raw := range []string{"mañana", "2027-02-30", "30/12/2027"} {
		task := &Task{Title: "Fecha mala"}
		task.SetDueDate(raw)

		verrs, err := ms.DB.ValidateAndCreate(task)
		ms.NoError(err)
		ms.True(verrs.HasAny())
		ms.Contains(verrs.Get("due_date"), "La fecha límite no es una fecha válida.")
		ms.Equal(raw, task.DueDateValue())
	}

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_DueDate_PastInvalidOnCreate() {
	for _, date := range []time.Time{daysFromToday(-1), daysFromToday(-365)} {
		task := &Task{Title: "Fecha pasada", DueDate: date}

		verrs, err := ms.DB.ValidateAndCreate(task)
		ms.NoError(err)
		ms.True(verrs.HasAny())
		ms.Contains(verrs.Get("due_date"), "La fecha límite no puede ser anterior a hoy.")
	}

	count, err := ms.DB.Count(&Task{})
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_Task_DueDate_TodayValidOnCreate() {
	today := Today()
	task := &Task{Title: "Vence hoy", DueDate: today}

	verrs, err := ms.DB.ValidateAndCreate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.Equal(today.Format(DueDateLayout), reloaded.DueDate.Format(DueDateLayout))
	ms.False(reloaded.PastDue())
}

func (ms *ModelSuite) Test_Task_DueDate_PastValidOnUpdate() {
	yesterday := daysFromToday(-1)
	task := &Task{Title: "Ya vencida", DueDate: yesterday}
	ms.NoError(ms.DB.Create(task))

	verrs, err := ms.DB.ValidateAndUpdate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	weekAgo := daysFromToday(-7)
	task.DueDate = weekAgo
	verrs, err = ms.DB.ValidateAndUpdate(task)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	reloaded := &Task{}
	ms.NoError(ms.DB.Find(reloaded, task.ID))
	ms.Equal(weekAgo.Format(DueDateLayout), reloaded.DueDate.Format(DueDateLayout))
}

func (ms *ModelSuite) Test_Task_Overdue() {
	ms.True(Task{DueDate: daysFromToday(-1)}.Overdue())
	ms.False(Task{DueDate: daysFromToday(0)}.Overdue())
	ms.False(Task{DueDate: daysFromToday(1)}.Overdue())

	completed := Task{DueDate: daysFromToday(-1), Completed: true}
	ms.False(completed.Overdue())
	ms.True(completed.PastDue())
}
