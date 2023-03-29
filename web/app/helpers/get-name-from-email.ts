import Helper from "@ember/component/helper";
import { inject as service } from "@ember/service";
import FetchService from "hermes/services/fetch";

export default class GetNameFromEmail extends Helper {
  @service("fetch") declare fetchSvc: FetchService;

  async compute([email]: [string]) {
    if (!email) {
      return "Unknown";
    }

    let fetchResponse = await this.fetchSvc.fetch(`/api/v1/people`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        query: email,
      }),
    });

    if (fetchResponse?.ok) {
      let result = await fetchResponse.json();
      return result[0]?.names[0].displayName || email;
    } else {
      return "Unknown";
    }
  }
}
