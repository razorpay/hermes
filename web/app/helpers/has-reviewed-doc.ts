import { helper } from "@ember/component/helper";
import { HermesDocument } from "hermes/types/document";

interface HasReviewedDocHelperSignature {
  Args: {
    Positional: [document: HermesDocument, approverEmail: string];
  };
  Return: boolean;
}

const hasReviewedDocHelper = helper<HasReviewedDocHelperSignature>(
  ([document, approverEmail]: [HermesDocument, string]) => {
    if (document.reviewedBy) {
      return document.reviewedBy.some((email) => email === approverEmail);
    } else {
      return false;
    }
  }
);

export default hasReviewedDocHelper;

declare module "@glint/environment-ember-loose/registry" {
  export default interface Registry {
    "has-reviewed-doc": typeof hasReviewedDocHelper;
  }
}
