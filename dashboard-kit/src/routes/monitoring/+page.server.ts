// src/routes/page-to-redirect/+page.server.ts

import { redirect } from '@sveltejs/kit';

export function load() {
    throw redirect(302, '/monitor');
}