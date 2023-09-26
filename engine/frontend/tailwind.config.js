/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
        "./src/**/*.{js,jsx,ts,tsx}",
    ],
    theme: {
        extend: {
            colors: {
                'kurtosis-background': '#171923',
            },
            minHeight: {
                '1/2': '50%',
            },
            minWidth: {
                '1/32': '3.125%',
            },
            Width: {
                '1/16': '6.25%',
            }
        },
    },
    plugins: [
        require('@tailwindcss/typography'),
    ],
}
