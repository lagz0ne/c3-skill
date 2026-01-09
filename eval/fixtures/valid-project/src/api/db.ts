import { Pool } from 'pg';

// Line 5 - db connection
export const db = new Pool({
  connectionString: process.env.DATABASE_URL,
});

export async function withTransaction<T>(
  fn: (client: any) => Promise<T>
): Promise<T> {
  const client = await db.connect();
  try {
    await client.query('BEGIN');
    const result = await fn(client);
    await client.query('COMMIT');
    return result;
  } catch (e) {
    await client.query('ROLLBACK');
    throw e;
  } finally {
    client.release();
  }
}
