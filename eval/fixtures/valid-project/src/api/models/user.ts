import { db } from '../db';

interface User {
  id: string;
  email: string;
  name: string;
  createdAt: Date;
}

// Line 8 - UserModel class
export class UserModel {
  async findAll(): Promise<User[]> {
    const result = await db.query('SELECT * FROM users');
    return result.rows;
  }

  async findById(id: string): Promise<User | null> {
    const result = await db.query('SELECT * FROM users WHERE id = $1', [id]);
    return result.rows[0] || null;
  }

  async create(data: { email: string; name: string }): Promise<User> {
    const result = await db.query(
      'INSERT INTO users (email, name) VALUES ($1, $2) RETURNING *',
      [data.email, data.name]
    );
    return result.rows[0];
  }

  async update(id: string, data: Partial<User>): Promise<User | null> {
    // Implementation
    return null;
  }

  async delete(id: string): Promise<boolean> {
    const result = await db.query('DELETE FROM users WHERE id = $1', [id]);
    return result.rowCount > 0;
  }
}
