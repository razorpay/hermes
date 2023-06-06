import Component from "@glimmer/component";
import {
  HdsButtonColor,
  HdsButtonIconPosition,
} from "hermes/types/hds-components";

interface XDropdownListToggleButtonComponentSignature {
  Element: HTMLButtonElement;
  Args: {
    registerAnchor: () => void;
    onTriggerKeydown: () => void;
    toggleContent: () => void;
    contentIsShown: boolean;
    disabled?: boolean;
    ariaControls: string;
    icon?: string;
    color?: HdsButtonColor;
    iconPosition?: HdsButtonIconPosition;
    text: string;
  };
  Blocks: {
    default: [];
  };
}

export default class XDropdownListToggleButtonComponent extends Component<XDropdownListToggleButtonComponentSignature> {}

declare module "@glint/environment-ember-loose/registry" {
  export default interface Registry {
    "x/dropdown-list/toggle-button": typeof XDropdownListToggleButtonComponent;
  }
}
