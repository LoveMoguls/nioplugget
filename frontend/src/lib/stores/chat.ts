import { writable } from 'svelte/store';

const API_BASE = typeof window !== 'undefined'
	? (import.meta.env.VITE_API_URL || 'http://localhost:8080')
	: 'http://localhost:8080';

export interface ChatMessage {
	id?: string;
	role: 'user' | 'assistant';
	content: string;
	createdAt?: string;
}

export const messages = writable<ChatMessage[]>([]);
export const isStreaming = writable(false);
export const sessionEnded = writable(false);

export async function sendMessage(sessionId: string, content: string): Promise<void> {
	// Add user message immediately
	messages.update((msgs) => [...msgs, { role: 'user', content }]);
	isStreaming.set(true);

	// Add empty assistant message for streaming
	messages.update((msgs) => [...msgs, { role: 'assistant', content: '' }]);

	try {
		const res = await fetch(`${API_BASE}/api/sessions/${sessionId}/messages`, {
			method: 'POST',
			credentials: 'include',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ content }),
		});

		if (!res.ok) throw new Error('Failed to send message');
		if (!res.body) throw new Error('No response body');

		const reader = res.body.getReader();
		const decoder = new TextDecoder();
		let buffer = '';

		while (true) {
			const { done, value } = await reader.read();
			if (done) break;

			buffer += decoder.decode(value, { stream: true });
			const lines = buffer.split('\n');
			buffer = lines.pop() || '';

			for (const line of lines) {
				if (line.startsWith('data: ')) {
					const data = line.slice(6).trim();
					if (data === '[DONE]') continue;
					try {
						const parsed = JSON.parse(data);
						if (parsed.text) {
							messages.update((msgs) => {
								const last = msgs[msgs.length - 1];
								if (last && last.role === 'assistant') {
									return [
										...msgs.slice(0, -1),
										{ ...last, content: last.content + parsed.text },
									];
								}
								return msgs;
							});
						}
						if (parsed.error) {
							messages.update((msgs) => {
								const last = msgs[msgs.length - 1];
								if (last && last.role === 'assistant') {
									return [...msgs.slice(0, -1), { ...last, content: parsed.error }];
								}
								return msgs;
							});
						}
					} catch {
						// Skip malformed JSON
					}
				}
			}
		}
	} catch {
		messages.update((msgs) => {
			const last = msgs[msgs.length - 1];
			if (last && last.role === 'assistant' && last.content === '') {
				return [...msgs.slice(0, -1), { ...last, content: 'Ett fel uppstod. Försök igen.' }];
			}
			return msgs;
		});
	} finally {
		isStreaming.set(false);
	}
}

export async function loadMessages(sessionId: string): Promise<void> {
	try {
		const res = await fetch(`${API_BASE}/api/sessions/${sessionId}/messages`, {
			credentials: 'include',
		});
		if (res.ok) {
			const data = await res.json();
			messages.set(
				data.map((m: { id: string; role: string; content: string; createdAt: string }) => ({
					id: m.id,
					role: m.role as 'user' | 'assistant',
					content: m.content,
					createdAt: m.createdAt,
				})),
			);
		}
	} catch {
		// Ignore - messages will be empty
	}
}

export function resetChat(): void {
	messages.set([]);
	isStreaming.set(false);
	sessionEnded.set(false);
}
