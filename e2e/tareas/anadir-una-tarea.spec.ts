import { test, expect, day } from '../fixtures';

test.describe('Tareas', () => {
  test('añadir una tarea', async ({ page }) => {
    // 1. La lista está vacía
    await expect(page.getByText('No hay tareas todavía.')).toBeVisible();

    // 2. Escribe el título, elige la fecha límite y pulsa «Añadir»
    await page.getByRole('textbox', { name: 'Nueva tarea' }).fill('Comprar pan');
    await page.getByLabel('Fecha límite').fill(day(1));
    await page.getByRole('button', { name: 'Añadir' }).click();

    // 3. La tarea aparece pendiente, con el aviso, la fecha y el contador
    await expect(page.getByRole('status')).toHaveText('Tarea creada.');
    await expect(page.locator('.task-title')).toHaveText(['Comprar pan']);
    await expect(page.getByRole('button', { name: 'Completar tarea: Comprar pan' })).toBeVisible();
    await expect(page.locator('.task-due time')).toHaveText(day(1).split('-').reverse().join('/'));
    await expect(page.getByText('Vencida')).toHaveCount(0);
    await expect(page.locator('.list-count')).toHaveText('1 tarea en tu lista');
  });
});
