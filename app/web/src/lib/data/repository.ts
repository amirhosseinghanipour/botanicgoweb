import { db, type LocalMessage, type LocalSession } from '$lib/db/indexeddb';
import { api } from '$lib/api/client';
import { API_URL } from '$lib/config';

// Toggle: local (IndexedDB) vs server
const PERSISTENCE_MODE: 'local' | 'server' = (import.meta as any).env?.VITE_PERSISTENCE_MODE || 'local';

export interface Repository {
  createSession(input: { title: string; model: string }): Promise<LocalSession>;
  getSessions(): Promise<LocalSession[]>;
  getSession(id: string): Promise<LocalSession | undefined>;
  deleteSession(id: string): Promise<void>;
  createMessage(input: { sessionId: string; role: 'user' | 'assistant'; content: string; model?: string }): Promise<LocalMessage>;
  getSessionMessages(sessionId: string): Promise<LocalMessage[]>;
}

function makeLocalRepository(): Repository {
  return {
    async createSession({ title, model }) {
      const now = new Date().toISOString();
      const session: LocalSession = {
        id: crypto.randomUUID(),
        title: title || 'New Chat',
        model,
        createdAt: now,
        updatedAt: now,
      };
      await db.sessions.put(session);
      return session;
    },
    async getSessions() {
      return db.sessions.orderBy('updatedAt').reverse().toArray();
    },
    async getSession(id) {
      return db.sessions.get(id);
    },
    async deleteSession(id) {
      await db.sessions.delete(id);
      // Also delete messages for that session
      await db.messages.where('sessionId').equals(id).delete();
    },
    async createMessage({ sessionId, role, content, model }) {
      const msg: LocalMessage = {
        id: crypto.randomUUID(),
        sessionId,
        role,
        content,
        model,
        createdAt: new Date().toISOString(),
      };
      await db.messages.put(msg);
      // bump session updatedAt
      const session = await db.sessions.get(sessionId);
      if (session) {
        session.updatedAt = msg.createdAt;
        await db.sessions.put(session);
      }
      return msg;
    },
    async getSessionMessages(sessionId) {
      return db.messages.where('sessionId').equals(sessionId).sortBy('createdAt');
    }
  };
}

function makeServerRepository(): Repository {
  return {
    async createSession({ title, model }) {
      const s = await api.createSession({ title, model });
      return { id: s.id, title: s.title, model: s.model, createdAt: s.created_at, updatedAt: s.updated_at } as LocalSession;
    },
    async getSessions() {
      const res = await api.getSessions();
      return res.map(s => ({ id: s.id, title: s.title, model: s.model, createdAt: s.created_at, updatedAt: s.updated_at }));
    },
    async getSession(id) {
      const s = await api.getSession(id);
      return s ? { id: s.id, title: s.title, model: (s as any).model || 'default', createdAt: s.created_at, updatedAt: s.updated_at } as LocalSession : undefined;
    },
    async deleteSession(id) {
      await api.deleteSession(id);
    },
    async createMessage({ sessionId, role, content, model }) {
      const m = await api.sendMessage(sessionId, content);
      return { id: m.id, sessionId, role, content: m.content, model: m.model, createdAt: m.created_at } as LocalMessage;
    },
    async getSessionMessages(sessionId) {
      const ms = await api.getSessionMessages(sessionId);
      return ms.map(m => ({ id: m.id, sessionId, role: m.role, content: m.content, model: m.model, createdAt: m.created_at }));
    }
  };
}

export const repository: Repository = PERSISTENCE_MODE === 'server' ? makeServerRepository() : makeLocalRepository();

