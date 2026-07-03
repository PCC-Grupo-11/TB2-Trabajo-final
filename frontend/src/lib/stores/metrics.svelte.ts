import { getToken } from '$lib/api/client';
import type { InfoMessage, MetricsMessage, WSMessage } from '$lib/types/metrics';

class MetricsStore {
	connected = $state(false);
	info = $state<InfoMessage | null>(null);
	metrics = $state<MetricsMessage | null>(null);

	private ws: WebSocket | null = null;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private reconnectDelay = 2000;

	connect() {
		if (this.ws) return;

		const token = getToken();
		if (!token) return;

		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const host = window.location.host;
		this.ws = new WebSocket(`${protocol}//${host}/ws/metrics?token=${token}`);

		this.ws.onopen = () => {
			this.connected = true;
			this.reconnectDelay = 2000;
		};

		this.ws.onmessage = (event) => {
			try {
				const msg: WSMessage = JSON.parse(event.data);
				if (msg.type === 'info') {
					this.info = msg;
				} else if (msg.type === 'metrics') {
					this.metrics = msg;
				}
			} catch {
				console.warn('Failed to parse WebSocket message:', event.data);
			}
		};

		this.ws.onclose = () => {
			this.connected = false;
			this.ws = null;
			this.scheduleReconnect();
		};

		this.ws.onerror = () => {
			this.ws?.close();
		};
	}

	disconnect() {
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}
		this.connected = false;
	}

	private scheduleReconnect() {
		if (this.reconnectTimer) return;
		this.reconnectTimer = setTimeout(() => {
			this.reconnectTimer = null;
			this.connect();
		}, this.reconnectDelay);
		this.reconnectDelay = Math.min(this.reconnectDelay * 1.5, 10000);
	}
}

export const metrics = new MetricsStore();
