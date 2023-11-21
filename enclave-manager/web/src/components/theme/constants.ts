import * as CSS from "csstype";

export const MAIN_APP_TOP_PADDING = "24px";
export const MAIN_APP_BOTTOM_PADDING = "20px";
export const MAIN_APP_LEFT_PADDING = "112px";
export const MAIN_APP_RIGHT_PADDING = "40px";

export const MAIN_APP_MAX_WIDTH: CSS.Property.MaxWidth | number = "1320px";
export const MAIN_APP_MAX_WIDTH_WITHOUT_PADDING: CSS.Property.MaxWidth | number = `${1320 - 112 - 40}px`;

export const MAIN_APP_PADDING: CSS.Property.Padding = `${MAIN_APP_TOP_PADDING} ${MAIN_APP_RIGHT_PADDING} ${MAIN_APP_BOTTOM_PADDING}  ${MAIN_APP_LEFT_PADDING}`;
export const MAIN_APP_TABPANEL_PADDING: CSS.Property.Padding = `0 ${MAIN_APP_RIGHT_PADDING} ${MAIN_APP_BOTTOM_PADDING}  ${MAIN_APP_LEFT_PADDING}`;
export const FLEX_STANDARD_GAP: CSS.Property.Gap | number = "32px";
