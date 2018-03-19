import React from 'react';

export default class InertiaClient {
    constructor(url) {
        this.url = "https://" + url;
    }

    /**
     * Makes a GET request to the given API endpoint with the given params and
     * returns the response in JSON format, or throws an error.
     * @param {String} endpoint
     * @param {Object} params
     */
    async _get(endpoint, params) {
        // @todo
    }

    /**
     * Makes a GET request to the given API endpoint with the given params and
     * returns the response in JSON format, or throws an error.
     * @param {String} endpoint 
     * @param {Object} params 
     */
    async _post(endpoint, params) {
        // @todo
    }
}
