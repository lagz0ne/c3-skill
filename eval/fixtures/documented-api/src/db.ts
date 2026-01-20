// Simple database abstraction
export const db = {
  async query(sql: string, params?: unknown[]) {
    // In real app, this would connect to MySQL/PostgreSQL
    console.log("Query:", sql, params);
    return [];
  }
};
