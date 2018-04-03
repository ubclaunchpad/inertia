import React from 'react';

export default class InertiaClient {
    constructor(url) {
        this.url = "https://" + url;
    }

    /**
     * Makes a GET request to the given API endpoint with the given params.
     * @param {String} endpoint
     * @param {Object} params
     */
    async _get(endpoint, params) {
        // @todo
    }

    /**
     * Makes a POST request to the given API endpoint with the given params.
     * @param {String} endpoint 
     * @param {Object} params 
     */
    async _post(endpoint, params) {
        const request = new Request(endpoint, params);

        try {
            return await fetch(request);
        } catch (e) {
            return e;
        }
    }
}
