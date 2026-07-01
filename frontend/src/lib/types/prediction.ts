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

export const TIME_BUCKET_SHORT = [
	'< 30 min',
	'< 1 h',
	'< 2.5 h',
	'< 7 h',
	'< 1 día',
	'< 3 días',
	'< 1 semana',
	'> 1 semana'
] as const;

export const TIME_BUCKET_FRIENDLY = [
	'Menos de 30 minutos',
	'Menos de 1 hora',
	'Menos de 2 horas y media',
	'Menos de 7 horas',
	'Menos de 1 día',
	'Menos de 3 días',
	'Menos de 1 semana',
	'Más de 1 semana'
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

export function getSeverityBorder(classIndex: number): string {
	const colors = [
		'border-prob-1',
		'border-prob-2',
		'border-prob-3',
		'border-prob-4',
		'border-prob-5',
		'border-prob-6',
		'border-prob-7',
		'border-prob-8'
	];
	return colors[classIndex] ?? 'border-prob-8';
}
