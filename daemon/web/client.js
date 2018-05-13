export default class InertiaClient {
  constructor(url) {
    this.url = 'https://' + url;
  }

  async logout() {
    const endpoint = '/user/logout';
    const params = {
      headers: {
        Accept: 'application/json',
      },
    };
    return this.post(endpoint, params);
  }

  async login(username, password) {
    const endpoint = '/user/login';
    const params = {
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: JSON.stringify({
        username,
        password,
      }),
    };
    return this.post(endpoint, params);
  }

  async validate() {
    return this.get('/user/validate', {});
  }

  async getContainerLogs(container = '/inertia-daemon') {
    const endpoint = '/logs';
    const params = {
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        stream: true,
        container,
      }),
    };
    return this.post(endpoint, params);
  }

  async getRemoteStatus() {
    const endpoint = '/status';
    const params = {
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
    };
    return this.get(endpoint, params);
  }

  /**
   * Makes a GET request to the given API endpoint with the given params.
   * @param {String} endpoint
   * @param {Object} params
   */
  async _get(endpoint, params) {
    const newParams = {
      ...params,
      method: 'GET',
      credentials: 'include',
    };

    const request = new Request(this.url + endpoint, newParams);

    try {
      return await fetch(request);
    } catch (e) {
      throw e;
    }
  }

  /**
   * Makes a POST request to the given API endpoint with the given params.
   * @param {String} endpoint
   * @param {Object} params
   */
  async _post(endpoint, params) {
    const newParams = {
      ...params,
      method: 'POST',
      credentials: 'include',
    };

    const request = new Request(this.url + endpoint, newParams);

    try {
      return await fetch(request);
    } catch (e) {
      throw e;
    }
  }
}
