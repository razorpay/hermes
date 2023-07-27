import Route from "@ember/routing/route";
import RSVP from "rsvp";
import { inject as service } from "@ember/service";
import timeAgo from "hermes/utils/time-ago";

export default class AuthenticatedProjectRoute extends Route {
    @service algolia;
    @service("config") configSvc;
    @service("fetch") fetchSvc;
    @service session;
    @service authenticatedUser;
  
    async model(params) {
      const projectId = params.project_id;
      const teamId = params.team_id;
      const businessUnitId = params.business_unit_id;
      console.log(projectId);
      console.log(teamId);
      console.log(businessUnitId);

      const userInfo = this.authenticatedUser.info;
      const searchIndex = this.configSvc.config.algolia_docs_index_name + "_createdTime_desc";
      const files = this.algolia.searchIndex
      .perform(searchIndex, "", {
        filters:
        `project:${projectId}` ,
        // `AND product:'${docid}'` +
        // `AND team:'${teamid}'` ,
        hitsPerPage: 1000,
      })
      .then((result) => {
        // Add modifiedAgo for each doc.
        for (const hit of result.hits) {
          this.fetchSvc
            .fetch("/api/v1/documents/" + hit.objectID)
            .then((resp) => resp.json())
            .then((doc) => {
              if (doc.modifiedTime) {
                const modifiedDate = new Date(doc.modifiedTime * 1000);
                hit.modifiedAgo = `Modified ${timeAgo(modifiedDate)}`;
              }
            })
            .catch((err) => {
              console.log(
                `Error getting document waiting for review (${hit.objectID}):`,
                err
              );
            });
        }
        return result.hits;
      });
      console.log(files);
      return RSVP.hash({
        files: files,
      });
    }
  }