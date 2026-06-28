import { predict } from '$lib/api/prediction';
import type { PredictRequest, PredictionResponse } from '$lib/types/prediction';
import { ApiRequestError } from '$lib/api/client';
import { getBorough } from '$lib/utils/borough';

function nycToUnix(dtStr: string): number {
    const [datePart, timePart] = dtStr.split(/[T ]/);
    const [year, month, day] = datePart.split('-').map(Number);
    const [hours, minutes] = timePart.split(':').map(Number);
    const now = new Date();
    const isDST = now.toLocaleString('en-US', { timeZone: 'America/New_York', timeZoneName: 'short' }).includes('EDT');
    return Math.floor(Date.UTC(year, month - 1, day, hours - (isDST ? -4 : -5), minutes) / 1000);
}

class PredictionStore {
    lat = $state<number | null>(null);
    lng = $state<number | null>(null);
    tsLocal = $state(new Date().toLocaleString('sv-SE', { timeZone: 'America/New_York' }).replace(' ', 'T').slice(0, 16));
    ts = $derived(nycToUnix(this.tsLocal));
    agency = $state('');
    complaintType = $state('');
    descriptor = $state('');
    locationType = $state('');
    borough = $state('');

    result = $state<PredictionResponse | null>(null);
    loading = $state(false);
    error = $state<string | null>(null);

    setCoords(newLat: number, newLng: number) {
        this.lat = newLat;
        this.lng = newLng;
        this.borough = getBorough(newLat, newLng);
        this.result = null;
        this.error = null;
    }

    async submitPrediction() {
        if (this.lat === null || this.lng === null) return;

        this.loading = true;
        this.error = null;

        try {
            const req: PredictRequest = {
                ts: this.ts,
                lat: this.lat,
                lon: this.lng,
                agency: this.agency,
                complaint_type: this.complaintType,
                descriptor: this.descriptor,
                location_type: this.locationType,
                borough: this.borough
            };
            this.result = await predict(req);
        } catch (err) {
            if (err instanceof ApiRequestError) {
                this.error = err.message || 'Error al predecir.';
            } else {
                this.error = 'Error de conexion.';
            }
        } finally {
            this.loading = false;
        }
    }
}

export const prediction = new PredictionStore();
