//helios.hashicorp.design/components/badge-count?tab=code#component-api

import { ComponentLike } from "@glint/template";
import {
  HdsBadgeCountColor,
  HdsBadgeCountSize,
  HdsBadgeCountType,
} from "hermes/types/hds-components";

export type HdsBadgeCountComponent = ComponentLike<{
  Element: HTMLDivElement;
  Args: {
    text: string;
    size?: HdsBadgeCountSize;
    type?: HdsBadgeCountType;
    color?: HdsBadgeCountColor;
  };
}>;
