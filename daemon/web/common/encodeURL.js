/**
 * Encodes the provided key-value pairs into an encoded
 * query string.
 *
 * >> encodeURL({tree: "bear", "penguin party": "igloo"});
 * >> // returns "tree=bear&penguin%20party=igloo"
 *
 * @param {{[key]: string}} params - An object containing
 * key-value pairs of the desired query params.
 * @returns {string} - Returns the encoded query string.
 */
function encodeURL(params) {
  const result = [];
  const keys = Object.keys(params);

  keys.forEach((k) => {
    const v = params[k];

    if (typeof v !== 'string') {
      throw new Error(`Arguments must be of type string, received [${k}, ${typeof v}]`);
    }

    result.push(encodeURIComponent(k) + '=' + encodeURIComponent(v));
  });

  return result.join('&');
}

module.exports = encodeURL;
