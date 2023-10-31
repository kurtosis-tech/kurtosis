import { Global } from "@emotion/react";

const Fonts = () => (
  <Global
    styles={`
      @font-face {
        font-family: "Inter";
        src: url("/fonts/Inter-VariableFont_slnt,wght.ttf") format("truetype-variations");
        font-weight: 125 950;
        font-stretch: 75% 125%;
        font-style: normal;
      }
      `}
  />
);

export default Fonts;
