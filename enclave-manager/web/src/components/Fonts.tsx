import { Global } from "@emotion/react";

const Fonts = () => (
  <Global
    styles={`
      @font-face {
        font-family: 'Gilroy';
        font-style: normal;
        font-weight: 400;
        font-display: swap;
        src: url('/fonts/Gilroy-Regular/font.woff2') format('woff2');
      },
      @font-face {
        font-family: 'Gilroy';
        font-style: bold;
        font-weight: 700;
        font-display: swap;
        src: url('/fonts/Gilroy-Bold/font.woff2') format('woff2');
      }
      `}
  />
)

export default Fonts