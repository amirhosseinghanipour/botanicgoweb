import Dexie, { type Table } from 'dexie';

export interface LocalSession {
  id: string;
  title: string;
  model: string;
  createdAt: string; // ISO
  updatedAt: string; // ISO
}

export interface LocalMessage {
  id: string;
  sessionId: string;
  role: 'user' | 'assistant';
  content: string;
  model?: string;
  createdAt: string; // ISO
}

export interface LocalSetting {
  key: string;
  value: unknown;
}

export class BotanicDB extends Dexie {
  sessions!: Table<LocalSession, string>;
  messages!: Table<LocalMessage, string>;
  settings!: Table<LocalSetting, string>;

  constructor() {
    super('botanic');
    this.version(1).stores({
      sessions: 'id, updatedAt, createdAt',
      messages: 'id, sessionId, createdAt',
      settings: 'key'
    });
  }
}

export const db = new BotanicDB();

