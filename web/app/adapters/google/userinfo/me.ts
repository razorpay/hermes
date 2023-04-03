import GoogleUserinfoAdapter from "../userinfo";

export default class GoogleUserinfoMeAdapter extends GoogleUserinfoAdapter {
  urlForQueryRecord() {
    let baseUrl = this.buildURL();
    return `${baseUrl}/me`;
  }
}
