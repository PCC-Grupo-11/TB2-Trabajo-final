import { apiClient } from '$lib/api/client';
import type { BulkPredictRequest, BulkPredictResponse } from '$lib/types/heatmap';

export async function bulkPredict(
	req: BulkPredictRequest,
	signal?: AbortSignal
): Promise<BulkPredictResponse> {
	return apiClient<BulkPredictResponse>('/predict/bulk', {
		method: 'POST',
		body: JSON.stringify(req),
		signal
	});
}
