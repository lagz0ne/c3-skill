export async function verify(token: string): Promise<{ sub: string }> {
  // In real app, this would verify JWT
  return { sub: "user-123" };
}

export async function sign(payload: Record<string, unknown>): Promise<string> {
  // In real app, this would sign JWT
  return "mock-token";
}
