import { writable } from 'svelte/store';

// Create a store to hold the fetched data
export const dataStore = writable<DashboardData | null>(null);