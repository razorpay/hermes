import Component from "@glimmer/component";
import ConfigService from "hermes/services/config";
import { inject as service } from "@ember/service";
import { HermesDocument } from "hermes/types/document";
import { tracked } from "@glimmer/tracking";
import { action } from "@ember/object";
import htmlElement from "hermes/utils/html-element";

interface DocumentSidebarHeaderComponentSignature {
  Args: {
    document: HermesDocument;
    isCollapsed: boolean;
    toggleCollapsed: () => void;
    userHasScrolled: boolean;
  };
}

export function isValidURL(input: string) {
  try {
    new URL(input);
    return true;
  } catch {
    return false;
  }
}

export default class DocumentSidebarHeaderComponent extends Component<DocumentSidebarHeaderComponentSignature> {
  @service("config") declare configSvc: ConfigService;

  @tracked protected shareModalIsShown = false;

  protected get modalContainer() {
    return htmlElement(".ember-application");
  }

  protected get shareButtonIsShown() {
    let { document } = this.args;
    return !document.isDraft && document.docNumber && document.docType;
  }

  protected get url() {
    let shortLinkBaseURL: string | undefined =
      this.configSvc.config.short_link_base_url;

    if (shortLinkBaseURL) {
      // Add a trailing slash if the URL needs one
      if (!shortLinkBaseURL.endsWith("/")) {
        shortLinkBaseURL += "/";
      }
      // Reject invalid URLs
      if (!isValidURL(shortLinkBaseURL)) {
        shortLinkBaseURL = undefined;
      }
    }
    return shortLinkBaseURL
      ? `${
          shortLinkBaseURL + this.args.document.docType.toLowerCase()
        }/${this.args.document.docNumber.toLowerCase()}`
      : window.location.href;
  }

  @action protected noop() {}

  @action selectText(event: FocusEvent) {
    event.preventDefault();
    let target = event.target as HTMLInputElement;
    target.select();
  }

  @action protected showShareModal() {
    this.shareModalIsShown = true;
  }

  @action protected hideShareModal() {
    this.shareModalIsShown = false;
  }
}
