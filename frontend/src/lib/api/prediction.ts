import { apiClient } from '$lib/api/client';
import type { PredictRequest, PredictionResponse } from '$lib/types/prediction';

export async function predict(req: PredictRequest): Promise<PredictionResponse> {
    console.log('[predict] request:', req);
    return apiClient<PredictionResponse>('/predict', {
        method: 'POST',
        body: JSON.stringify(req)
    });
}
