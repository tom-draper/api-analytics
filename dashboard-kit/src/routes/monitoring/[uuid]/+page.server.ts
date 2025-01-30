// src/routes/page-to-redirect/+page.server.ts

import { page } from '$app/state';
import { redirect } from '@sveltejs/kit';

export function load() {
    const uuid = page.params.uuid;
    console.log(page)
    if (!uuid) {
        throw redirect(308, '/monitor');
    }

    throw redirect(308, `/monitor/${page.params.uuid}`);
}