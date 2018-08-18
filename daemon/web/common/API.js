import encodeURL from './encodeURL';

const api = process.env.INERTIA_API || '';

export default class InertiaAPI {
  static async logout() {
    const endpoint = '/user/logout';
    const params = {
      headers: {
        Accept: 'application/json',
      },
    };

    const resp = await InertiaAPI.post(endpoint, params);
    const body = await resp.json();
    switch (resp.status) {
      case 200:
        return body;
      default:
        throw new Error(
          `logout failed with status ${resp.status}: ${JSON.stringify(body)}`
        );
    }
  }

  static async login(username, password) {
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

    const resp = await InertiaAPI.post(endpoint, params);
    const body = await resp.json();
    switch (resp.status) {
      case 200:
        return body;
      default:
        throw new Error(
          `login failed with status ${resp.status}: ${JSON.stringify(body)}`
        );
    }
  }

  static async validate() {
    return InertiaAPI.get('/user/validate', {});
  }

  static async getContainerLogs(container = '/inertia-daemon') {
    const endpoint = '/logs';
    const queryParams = {
      stream: 'true',
      container,
    };
    const params = {
      headers: {
        'Content-Type': 'application/json',
      },
    };

    // todo: websockets
    return InertiaAPI.post(endpoint, params, queryParams);
  }

  static async getRemoteStatus() {
    const endpoint = '/status';
    const params = {
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
    };

    const resp = await InertiaAPI.post(endpoint, params);
    const body = await resp.json();
    switch (resp.status) {
      case 200:
        return body;
      default:
        throw new Error(
          `status check failed with status ${resp.status}: ${JSON.stringify(body)}`
        );
    }
  }

  /**
   * Makes a GET request to the given API endpoint with the given params.
   * @param {String} endpoint
   * @param {Object} params
   * @param {{[key]: string}} queryParams
   */
  static async get(endpoint, params, queryParams) {
    const newParams = {
      ...params,
      method: 'GET',
      credentials: 'include',
    };
    const queryString = queryParams ? encodeURL(queryParams) : '';
    const url = api + endpoint + queryString;
    return fetch(new Request(url, newParams));
  }

  /**
   * Makes a POST request to the given API endpoint with the given params.
   * @param {String} endpoint
   * @param {Object} params
   */
  static async post(endpoint, params) {
    const newParams = {
      ...params,
      method: 'POST',
      credentials: 'include',
    };
    return fetch(new Request(api + endpoint, newParams));
  }
}
