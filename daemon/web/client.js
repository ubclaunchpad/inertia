import React from 'react';
import encodeURL from './common/encodeURL';

export default class InertiaClient {
  constructor(url) {
    this.url = 'https://' + url;
  }

  async logout() {
    const endpoint = '/user/logout';
    const params = {
      headers: {
        'Accept': 'application/json'
      }
    };
    return this._post(endpoint, params);
  }

  async login(username, password) {
    const endpoint = '/user/login';
    const params = {
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/x-www-form-urlencoded'
      },
      body: JSON.stringify({
        username: username,
        password: password,
      })
    };
    return this._post(endpoint, params);
  }

  async validate() {
    return this._get('/user/validate', {});
  }

  async getContainerLogs(container = '/inertia-daemon') {
    const endpoint = '/logs';
    const queryParams = {
      stream: 'true',
      container: container,
    };
    const params = {
      headers: {
        'Content-Type': 'application/json',
      }
    };

    return this._post(endpoint, params, queryParams);
  }

  async getRemoteStatus() {
    const endpoint = '/status';
    const params = {
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      }
    };
    return this._get(endpoint, params);
  }

  /**
   * Makes a GET request to the given API endpoint with the given params.
   * @param {String} endpoint
   * @param {Object} params
   * @param {{[key]: string}} queryParams
   */
  async _get(endpoint, params, queryParams) {
    const queryString = queryParams ? encodeURL(queryParams) : '';
    const url = this.url + endpoint + queryString;

    params.method = 'GET';
    params.credentials = 'include';

    const request = new Request(url, params);

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
    params.method = 'POST';
    params.credentials = 'include';
    const request = new Request(this.url + endpoint, params);
    try {
      return await fetch(request);
    } catch (e) {
      throw e;
    }
  }
}
