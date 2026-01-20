import { Request, Response, NextFunction } from 'express';

export function authMiddleware(req: Request, res: Response, next: NextFunction) {
  const token = req.headers.authorization?.replace('Bearer ', '');

  if (!token) {
    return res.status(401).json({ error: 'No token provided' });
  }

  // Simple token validation (in real app, verify JWT)
  if (token === 'invalid') {
    return res.status(401).json({ error: 'Invalid token' });
  }

  next();
}
