import Component from "@glimmer/component";
import { dasherize } from "@ember/string";
import getProductId from "hermes/utils/get-product-id";

interface DocThumbnailComponentSignature {
  Element: HTMLDivElement;
  Args: {
    isLarge?: boolean;
    status?: string;
    product?: string;
    docID?: string;
  };
}

export default class DocThumbnailComponent extends Component<DocThumbnailComponentSignature> {
  protected get status(): string | null {
    if (this.args.status) {
      return dasherize(this.args.status);
    } else {
      return null;
    }
  }

  protected get productShortName(): string | null {
    if (this.args.product) {
      return getProductId(this.args.product);
    } else {
      return null;
    }
  }

  protected get isReviewed(): boolean {
    return this.status === "reviewed";
  }

  protected get isObsolete(): boolean {
    return this.status === "obsolete";
  }
}

declare module "@glint/environment-ember-loose/registry" {
  export default interface Registry {
    "Doc::Thumbnail": typeof DocThumbnailComponent;
  }
}
