import Component from "@glimmer/component";
import { HdsComponentSize } from "hermes/enums/hds-components";

interface XDropdownListCheckableItemComponentSignature {
  Args: {
    selected: boolean;
    value: string;
    count?: number;
  };
}

export default class XDropdownListCheckableItemComponent extends Component<XDropdownListCheckableItemComponentSignature> {
  HdsBadgeCountSize = HdsComponentSize;
}

declare module "@glint/environment-ember-loose/registry" {
  export default interface Registry {
    "X::DropdownList::CheckableItem": typeof XDropdownListCheckableItemComponent;
  }
}
