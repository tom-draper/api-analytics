<script lang="ts">
    import { onMount } from 'svelte';
    import periodToDays from '../../../lib/period';
    import { CREATED_AT, STATUS } from '../../../lib/consts';
    import type { Period } from '../../../lib/settings';

    function daysAgo(date: Date): number {
        let now = new Date();
        return Math.floor(
            (now.getTime() - date.getTime()) / (24 * 60 * 60 * 1000)
        );
    }

    function setSuccessRate() {
        let success = {};
        let minDate = Number.POSITIVE_INFINITY;
        for (let i = 1; i < data.length; i++) {
            let date = new Date(data[i][CREATED_AT]);
            date.setHours(0, 0, 0, 0);
            let dateStr = date.toDateString();
            if (!(dateStr in success)) {
                success[dateStr] = { total: 0, successful: 0 };
            }
            if (data[i][STATUS] >= 200 && data[i][STATUS] <= 299) {
                success[dateStr].successful++;
            }
            success[dateStr].total++;
            if ((date as any) < minDate) {
                minDate = date as any;
            }
        }

        let days = periodToDays(period);
        if (days == null) {
            days = daysAgo(minDate as any);
        }

        let successArr = new Array(days).fill(-0.1); // -0.1 -> 0
        for (let date in success) {
            let idx = daysAgo(new Date(date));
            successArr[successArr.length - 1 - idx] =
                success[date].successful / success[date].total;
        }

        successRate = successArr;
    }

    function build() {
        setSuccessRate();
    }

    let successRate: any[];
    let mounted = false;
    onMount(() => {
        mounted = true;
    });

    $: successRate;

    $: data && mounted && build();

    export let data: RequestsData, period: Period;
</script>

<div class="success-rate-container">
    {#if successRate != undefined}
        <div class="success-rate-title">Success rate</div>
        <div class="errors">
            {#each successRate as value}
                <div
                    class="error level-{Math.floor(value * 10) + 1}"
                    title="Success rate: {value >= 0
                        ? (value * 100).toFixed(1) + '%'
                        : 'No requests'}"
                />
            {/each}
        </div>
    {/if}
</div>

<style>
    .errors {
        display: flex;
        margin-top: 8px;
        margin: 0 10px 0 35px;
    }
    .error {
        background: var(--highlight);
        flex: 1;
        height: 40px;
        margin: 0 0.1%;
        border-radius: 1px;
    }
    .success-rate-container {
        text-align: left;
        font-size: 0.9em;
        color: var(--dim-text);
    }
    .success-rate-title {
        margin: 0 0 4px 43px;
    }
    .success-rate-container {
        margin: 1.5em 2.5em 2em;
    }
    .level-0 {
        background: rgb(40, 40, 40);
    }
    .level-1 {
        background: #e46161;
    }
    .level-2 {
        background: #f18359;
    }
    .level-3 {
        background: #f5a65a;
    }
    .level-4 {
        background: #f3c966;
    }
    .level-5 {
        background: #ebeb81;
    }
    .level-6 {
        background: #c7e57d;
    }
    .level-7 {
        background: #a1df7e;
    }
    .level-8 {
        background: #77d884;
    }
    .level-9 {
        background: #3fcf8e;
    }
</style>
