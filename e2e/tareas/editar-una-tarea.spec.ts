import { test, expect, day } from '../fixtures';

test.describe('Tareas', () => {
  test('editar una tarea', async ({ page, tasks }) => {
    await tasks([{ title: 'Llamar al proveedor' }, { title: 'Revisar el presupuesto' }]);
    await page.reload();

    // 1. Abre la edición de «Llamar al proveedor»: el campo trae la fecha guardada
    await page.getByRole('link', { name: 'Editar tarea: Llamar al proveedor' }).click();
    await expect(page.getByRole('heading', { name: 'Editar tarea' })).toBeVisible();
    await expect(page.getByLabel('Fecha límite')).toHaveValue(day(0));

    // 2. Cambia el título y la fecha y guarda
    await page.getByRole('textbox', { name: 'Título' }).fill('Llamar al proveedor de papel');
    await page.getByLabel('Fecha límite').fill(day(7));
    await page.getByRole('button', { name: 'Guardar cambios' }).click();

    // 3. Vuelve a la lista con el título y la fecha nuevos, en el mismo sitio
    await expect(page.getByRole('status')).toHaveText('Tarea actualizada.');
    await expect(page.locator('.task-title')).toHaveText(['Llamar al proveedor de papel', 'Revisar el presupuesto']);
    await expect(page.locator('.task-due time').first()).toHaveText(day(7).split('-').reverse().join('/'));
  });
});
