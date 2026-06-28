export interface PredictRequest {
	ts: number;
	lat: number;
	lon: number;
	agency: string;
	complaint_type: string;
	descriptor: string;
	location_type: string;
	borough: string;
}

export interface PredictionResponse {
	class: number;
	confidence: number;
	probabilities: number[];
	latency_ms: number;
	cached: boolean;
}

export const TIME_BUCKETS = [
	'< 30 minutos',
	'< 1 hora',
	'< 2 horas 30 minutos',
	'< 7 horas',
	'< 1 día',
	'< 3 días',
	'< 1 semana',
	'> 1 semana'
] as const;

export function getSeverityColor(classIndex: number): string {
	if (classIndex <= 1) return 'bg-severity-fast';
	if (classIndex <= 3) return 'bg-severity-moderate';
	if (classIndex <= 5) return 'bg-severity-slow';
	return 'bg-severity-critical';
}

export function getSeverityLabel(confidence: number): string {
	if (confidence > 0.7) return 'Alta';
	if (confidence > 0.4) return 'Media';
	return 'Baja';
}
