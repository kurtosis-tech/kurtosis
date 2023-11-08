import { Global } from "@emotion/react";

/*
 * Source: https://fonts.googleapis.com/css2?family=Inter:wght@500&display=swap
 * */
const Fonts = () => (
  <Global
    styles={`
        @font-face {
          font-family: 'Inter';
          font-stretch: 75% 125%;
          font-style: normal;
          font-weight: 500;
          font-display: swap;
          src: url(https://fonts.gstatic.com/s/inter/v13/UcCO3FwrK3iLTeHuS_fvQtMwCp50KnMw2boKoduKmMEVuI6fAZJhiJ-Ek-_EeAmM.woff2) format('woff2');
          unicode-range: U+0460-052F, U+1C80-1C88, U+20B4, U+2DE0-2DFF, U+A640-A69F, U+FE2E-FE2F;
        }
      `}
  />
);

export default Fonts;
