<script>
    /** @type {string} */
    export let size = '200px';
    /** @type {string} */
    export let className = '';
</script>

<div class="relative flex items-center justify-center {className}" style="width: {size}; height: {size};">
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" preserveAspectRatio="xMidYMid" class="w-full h-full">
        <circle cx="50" cy="50" fill="none" stroke="#e2e8f0" stroke-width="8" r="35" stroke-dasharray="164.93361431346415 56.97787143782138">
        </circle>
        <circle cx="50" cy="50" fill="none" stroke="#3b82f6" stroke-width="8" r="35" stroke-dasharray="164.93361431346415 56.97787143782138" stroke-linecap="round">
            <animateTransform attributeName="transform" type="rotate" repeatCount="indefinite" dur="1s" values="0 50 50;360 50 50" keyTimes="0;1"></animateTransform>
        </circle>
    </svg>
</div>
