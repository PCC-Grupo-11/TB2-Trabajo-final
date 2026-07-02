export interface BulkPredictRequest {
	ts: number;
	agency: string;
	complaint_type: string;
	descriptor: string;
	location_type: string;
	hexes: Record<string, string[]>;
}

export interface HexPrediction {
	hex: string;
	class: number;
	confidence: number;
}

export interface BulkPredictResponse {
	results: HexPrediction[];
	latency_ms: number;
	cached: boolean;
}

export interface HeatmapViewport {
	north: number;
	south: number;
	east: number;
	west: number;
}
