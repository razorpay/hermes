import Component from "@glimmer/component";
import { A } from "@ember/array";
import { action } from "@ember/object";
import { tracked } from "@glimmer/tracking";
import { inject as service } from "@ember/service";
import AlgoliaService from "hermes/services/algolia";
import { HermesDocument } from "hermes/types/document";
import FetchService from "hermes/services/fetch";
import { dropTask, restartableTask } from "ember-concurrency";
import NativeArray from "@ember/array/-private/native-array";
import ConfigService from "hermes/services/config";
import FlashMessageService from "ember-cli-flash/services/flash-messages";
import { next } from "@ember/runloop";
import { assert } from "@ember/debug";

interface InputsDocumentSelect3ComponentSignature {
  Args: {
    productArea?: string;
    objectID?: string;
  };
}

export interface RelatedExternalLink {
  url: string;
  displayURL: string;
}

// const GOOGLE_FAVICON_URL_PREFIX =
//   "https://s2.googleusercontent.com/s2/favicons";

export default class InputsDocumentSelect3Component extends Component<InputsDocumentSelect3ComponentSignature> {
  @service("config") declare configSvc: ConfigService;
  @service("fetch") declare fetchSvc: FetchService;
  @service declare algolia: AlgoliaService;
  @service declare flashMessages: FlashMessageService;

  @tracked query = "";
  @tracked inputValueIsValid = false;
  @tracked relatedLinks: NativeArray<RelatedExternalLink> = A();
  @tracked relatedDocuments: NativeArray<HermesDocument> = A();
  @tracked faviconURL: string | null = null;
  @tracked _shownDocuments: HermesDocument[] | null = null;
  @tracked searchInput: HTMLInputElement | null = null;

  @tracked modalIsShown = false;

  get relatedResources(): NativeArray<RelatedExternalLink | HermesDocument> {
    let resources: NativeArray<RelatedExternalLink | HermesDocument> = A();
    resources.pushObjects(this.relatedDocuments);
    resources.pushObjects(this.relatedLinks);
    return resources;
  }

  get shownDocuments(): { [key: string]: HermesDocument } {
    /**
     * The array initially looks like this:
     * [{title: "foo", objectID: "bar"...}, ...]
     *
     * We need it to look like:
     * { "bar": {title: "foo", objectID: "bar"...}, ...}
     */

    let documents: any = {};

    if (this._shownDocuments) {
      this._shownDocuments.forEach((doc) => {
        documents[doc.objectID] = doc;
      });
    }
    return documents;
  }

  @action showModal() {
    this.modalIsShown = true;
  }

  @action hideModal() {
    this.modalIsShown = false;
  }

  @action addRelatedExternalLink() {
    let displayURL;

    if (this.query.startsWith("http://")) {
      displayURL = this.query.replace("http://", "");
    } else if (this.query.startsWith("https://")) {
      displayURL = this.query.replace("https://", "");
    } else {
      displayURL = this.query;
    }

    if (displayURL.startsWith("www.")) {
      displayURL = displayURL.replace("www.", "");
    }

    if (displayURL.endsWith("/")) {
      displayURL = displayURL.slice(0, -1);
    }

    let externalLink = {
      url: this.query,
      displayURL: displayURL,
    };

    const isDuplicate = this.relatedLinks.find((link) => {
      return link.url === externalLink.url;
    });

    if (isDuplicate) {
      this.showDuplicateMessage();
    } else {
      this.relatedLinks.unshiftObject(externalLink);
    }
  }

  @action addRelatedDocument(documentObjectID: string) {
    let document = this.shownDocuments[documentObjectID];
    if (document) {
      this.relatedDocuments.unshiftObject(document);
    }
    this.hideModal();
  }

  @action showDuplicateMessage() {
    this.flashMessages.add({
      title: "Duplicate URL",
      message: "This link is already a related doc.",
      type: "critical",
      timeout: 6000,
      extendedTimeout: 1000,
    });
  }

  @action onInputKeydown(dd: any, e: KeyboardEvent) {
    if (e.key === "Enter") {
      if (this.inputValueIsValid) {
        this.addRelatedExternalLink();
        dd.hideContent();
      }
    }
    if (e.key === "ArrowUp" || e.key === "ArrowDown") {
      if (!this.query.length) {
        e.preventDefault();
        return;
      }
    }

    dd.onTriggerKeydown(dd.contentIsShown, dd.showContent, e);
  }

  @action didInsertInput(dd: any, e: HTMLInputElement) {
    this.searchInput = e;
    void this.loadInitialData.perform(dd);

    next(() => {
      assert("searchInput expected", this.searchInput);
      this.searchInput.focus();
    });
  }

  @action maybeCloseDropdown(dd: any) {
    if (dd.contentIsShown) {
      dd.hideContent();
    }
  }

  @action maybeOpenDropdown(dd: any) {
    if (!dd.contentIsShown) {
      dd.showContent();
    }
  }

  @action removeResource(resource: RelatedExternalLink | HermesDocument) {
    // if the resource is a RelatedExternalLink, remove it from the relatedLinks array
    // otherwise, remove it from the relatedDocuments array

    if ("displayURL" in resource) {
      this.relatedLinks.removeObject(resource);
      return;
    } else {
      this.relatedDocuments.removeObject(resource);
      return;
    }
  }

  protected loadInitialData = dropTask(async (dd: any) => {
    await this.search.perform(dd, "");
  });

  protected search = restartableTask(async (dd: any, query: string) => {
    let index =
      this.configSvc.config.algolia_docs_index_name +
      "_createdTime_desc__productRanked";

    let filterString = `(NOT objectID:"${this.args.objectID}")`;

    if (this.relatedDocuments.length) {
      let relatedDocIDs = this.relatedDocuments.map((doc) => doc.objectID);

      filterString = filterString.slice(0, -1) + " ";

      filterString += `AND NOT objectID:"${relatedDocIDs.join(
        '" AND NOT objectID:"'
      )}")`;
    }

    try {
      let algoliaResponse = await this.algolia.searchIndex
        .perform(index, query, {
          hitsPerPage: 5,
          filters: filterString,
          attributesToRetrieve: ["title", "product", "docNumber"],
          optionalFilters: ["product:" + this.args.productArea],
        })
        .then((response) => response);
      if (algoliaResponse) {
        this._shownDocuments = algoliaResponse.hits as HermesDocument[];
        dd.resetFocusedItemIndex();
      }

      next(() => {
        dd.scheduleAssignMenuItemIDs();
      });
    } catch (e) {
      console.error(e);
    }
  });

  @action protected onInput(dd: any, e: Event) {
    const input = e.target as HTMLInputElement;
    this.query = input.value;

    void this.search.perform(dd, this.query);
  }
}
