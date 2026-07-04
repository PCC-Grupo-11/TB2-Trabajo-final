export interface InfoNode {
	addr: string;
	cpu_name: string;
	cores: number;
	total_ram_gb: number;
}

export interface InfoMessage {
	type: 'info';
	api: InfoNode;
	nodes: InfoNode[];
}

export interface NodeMetrics {
	addr: string;
	status: string;
	cpu_percent: number;
	memory_bytes: number;
	requests_served: number;
}

export interface MetricsMessage {
	type: 'metrics';
	api: {
		cpu_percent: number;
		memory_bytes: number;
	};
	nodes: NodeMetrics[];
	cluster: {
		total_predictions: number;
		avg_latency_ms: number;
		cache_hit_rate: number;
		total_nodes: number;
		healthy_nodes: number;
	};
}

export type WSMessage = InfoMessage | MetricsMessage;

export interface HistoryPoint {
	ts: Date;
	api: { cpu: number; mem: number };
	nodes: Record<string, { cpu: number; mem: number; req: number }>;
}
