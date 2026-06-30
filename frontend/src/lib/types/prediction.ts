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
	const colors = [
		'bg-prob-1',
		'bg-prob-2',
		'bg-prob-3',
		'bg-prob-4',
		'bg-prob-5',
		'bg-prob-6',
		'bg-prob-7',
		'bg-prob-8'
	];
	return colors[classIndex] ?? 'bg-prob-8';
}

export function getSeverityLabel(confidence: number): string {
	if (confidence > 0.7) return 'Alta';
	if (confidence > 0.4) return 'Media';
	return 'Baja';
}
