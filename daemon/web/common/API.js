import encodeURL from './encodeURL';

export default class InertiaAPI {
  static async logout() {
    const endpoint = '/user/logout';
    const params = {
      headers: {
        Accept: 'application/json',
      },
    };
    return InertiaAPI.post(endpoint, params);
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
    return InertiaAPI.post(endpoint, params);
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
    return InertiaAPI.get(endpoint, params);
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
    const url = endpoint + queryString;

    const request = new Request(url, newParams);

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
  static async post(endpoint, params) {
    const newParams = {
      ...params,
      method: 'POST',
      credentials: 'include',
    };

    const request = new Request(endpoint, newParams);

    try {
      return await fetch(request);
    } catch (e) {
      throw e;
    }
  }
}
