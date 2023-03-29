import { action } from "@ember/object";
import { inject as service } from "@ember/service";
import Component from "@glimmer/component";
import { tracked } from "@glimmer/tracking";
import { task } from "ember-concurrency";
import AlgoliaService from "hermes/services/algolia";
import ConfigService from "hermes/services/config";
import { HermesDocument } from "hermes/types/document";
import { SearchResponse } from "instantsearch.js";

// @ts-ignore - not yet typed
import timeAgo from "hermes/utils/time-ago";
import FetchService from "hermes/services/fetch";

interface DashboardLatestUpdatesComponentSignature {
  Args: {};
}

export default class DashboardLatestUpdatesComponent extends Component<DashboardLatestUpdatesComponentSignature> {
  @service("config") declare configSvc: ConfigService;
  @service("fetch") declare fetchSvc: FetchService;

  @service declare algolia: AlgoliaService;

  @tracked currentTab = "new";
  @tracked docsToShow: HermesDocument[] | null = null;

  /**
   * The message to show when there are no docs for a given tab.
   */
  get emptyStateMessage() {
    switch (this.currentTab) {
      case "new":
        return "No documents have been created yet.";
      case "in-review":
        return "No docs are in review.";
      case "approved":
        return "No docs have been approved.";
    }
  }
  /**
   * Calls the initial fetchDocs task.
   * Used in the template to show a loader on initial load.
   */
  didInsert = task(async () => {
    await this.fetchDocs.perform();
  });

  /**
   * Set the current tab (if necessary) and fetch its docs.
   */
  @action setCurrentTab(tab: string) {
    if (tab !== this.currentTab) {
      this.currentTab = tab;
      this.fetchDocs.perform();
    }
  }

  /**
   * Sends an Algolia query to fetch the docs for the current tab.
   * Called onLoad and when tabs are changed.
   */
  fetchDocs = task(async () => {
    let facetFilters = "";

    // Translate the current tab to an Algolia facetFilter.
    switch (this.currentTab) {
      case "new":
        facetFilters = "";
        break;
      case "in-review":
        facetFilters = "status:In-Review";
        break;
      case "approved":
        facetFilters = "status:approved";
        break;
    }

    await this.algolia.clearCache.perform();

    let newDocsToShow = await this.algolia.searchIndex
      .perform(
        this.configSvc.config.algolia_docs_index_name + "_modifiedTime_desc",
        "",
        {
          facetFilters: [facetFilters],
          hitsPerPage: 4,
        }
      )
      .then((result: SearchResponse<unknown>) => {
        // Add modifiedAgo for each doc.
        for (const hit of result.hits as HermesDocument[]) {
          if (hit.modifiedTime) {
            const modifiedAgo = new Date(hit.modifiedTime * 1000);
            hit.modifiedAgo = `Modified ${timeAgo(modifiedAgo)}`;
          }
        }
        return result.hits as HermesDocument[];
      });

    // need to loop through each docsToShow and fetch the owner name from the email.
    // the fetch requests must happen in parallel, so we use Promise.all
    if (newDocsToShow) {
      this.docsToShow = await Promise.all(
        newDocsToShow.map(async (doc) => {
          let response = await this.fetchSvc
            .fetch("/api/v1/people", {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify({
                query: doc.owners[0],
              }),
            })
            .then((response) => response?.json());

          let name = response[0]?.names[0].displayName;
          if (name) {
            doc.owners[0] = name;
          }

          return doc;
        })
      );
    }
  });
}
