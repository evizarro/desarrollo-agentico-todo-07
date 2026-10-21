import { test, expect, day } from '../fixtures';

test.describe('Tareas', () => {
  test('«Vencida» al completar y reabrir', async ({ page, tasks }) => {
    await tasks([
      { title: 'Preparar la reunión', dueDate: day(1) },
      { title: 'Pagar la factura', dueDate: day(-1) },
      { title: 'Llamar al gestor', dueDate: day(0) },
      { title: 'Renovar el dominio', dueDate: day(-1), completed: true },
    ]);
    await page.reload();

    const row = page.getByRole('listitem').filter({ hasText: 'Pagar la factura' });

    // 1. Solo la pendiente con fecha anterior a hoy muestra «Vencida»
    await expect(page.getByText('Vencida')).toHaveCount(1);
    await expect(row.getByText('Vencida')).toBeVisible();

    // 2. Al completarla, sin recargar, deja de estar vencida
    await page.getByRole('button', { name: 'Completar tarea: Pagar la factura' }).click();
    await expect(row.getByText('Vencida')).toHaveCount(0);

    // 3. Al reabrirla, sin recargar, vuelve a estarlo
    await page.getByRole('button', { name: 'Reabrir tarea: Pagar la factura' }).click();
    await expect(row.getByText('Vencida')).toBeVisible();

    // 4. Tras recargar sigue marcada
    await page.reload();
    await expect(row.getByText('Vencida')).toBeVisible();
  });
});
