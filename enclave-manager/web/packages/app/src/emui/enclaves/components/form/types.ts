import * as CSS from "csstype";
import { FieldPath } from "react-hook-form";

export interface KurtosisFormInputProps<DataModel extends object> {
  name: FieldPath<DataModel>;
  placeholder?: string;
  isRequired?: boolean;
  validate?: (value: any) => string | undefined;
  disabled?: boolean;
  width?: CSS.Property.Width;
  size?: string;
  tabIndex?: number;
}
