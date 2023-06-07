// https://helios.hashicorp.design/components/form/text-input?tab=code#formtextinputbase-1
import { ComponentLike } from "@glint/template";
import { HdsBadgeCountSize } from "hermes/types/HdsBadgeCountSize";

export type HdsFormTextInputBase = ComponentLike<{
  Element: HTMLInputElement;
  Args: {
    type: string;
    value: string | number | Date;
    isInvalid?: boolean;
    width?: string;
  };
}>;
