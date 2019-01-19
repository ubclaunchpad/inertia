import {
  GET_PROJECT_DETAILS_SUCCESS,
  GET_PROJECT_DETAILS_FAILURE,
  GET_PROJECT_LOGS_SUCCESS,
  GET_PROJECT_LOGS_FAILURE,
  TEST,
} from './_constants';

import {
  MOCK_DETAILS,
  MOCK_LOGS,
} from './_mock';

import api from '../api';

function promiseState(p) {
  const t = {};

  return Promise.race([p, t])
    .then(v => (v === t ? 'pending' : ('fulfilled', () => 'rejected')));
}

function detailsDispatch(payload, failure = false) {
  return {
    type: failure ? GET_PROJECT_DETAILS_SUCCESS : GET_PROJECT_DETAILS_FAILURE,
    payload,
  };
}

export const getStatus = () => {
  if (TEST) return (dispatch) => {
    // TODO: put fetch request here
    dispatch({
      type: GET_PROJECT_DETAILS_SUCCESS,
      payload: { project: MOCK_DETAILS },
    });
  };

  return async (dispatch) => {
    try {
      const status = await api.getRemoteStatus();
      dispatch(detailsDispatch({ status }));
    } catch (error) {
      dispatch(detailsDispatch({ error }, true));
    }
  };
};

export const getLogs = ({ container }) => (dispatch) => {
  try {
    let resp;
    if (!container) {
      resp = MOCK_LOGS;
      // resp = InertiaAPI.getContainerLogs();
    } else {
      resp = MOCK_LOGS;
      // resp = InertiaAPI.getContainerLogs(container);
    }
    if (resp.status !== 200) {
      // TODO: error dispatch here
    }

    const reader = resp.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let buffer = '';
    const stream = () => promiseState(reader.closed)
      .then((s) => {
        if (s === 'pending') {
          return reader.read()
            .then((data) => {
              const chunk = decoder.decode(data.value);
              const parts = chunk.split('\n')
                .filter(c => c);

              parts[0] = buffer + parts[0];
              buffer = '';
              if (!chunk.endsWith('\n')) {
                buffer = parts.pop();
              }

              // TODO: concatenate logs and add to dispatch
              dispatch({
                type: GET_PROJECT_LOGS_SUCCESS,
                payload: { logs: MOCK_LOGS },
              });

              return stream();
            });
        }
        return null;
      })
      .catch(() => {
        dispatch({
          // TODO: change to failure AC
          type: GET_PROJECT_LOGS_SUCCESS,
          payload: { logs: MOCK_LOGS } });
      });

    stream();
  } catch (e) {
    dispatch({
      type: GET_PROJECT_LOGS_FAILURE,
      payload: { logs: MOCK_LOGS } });
  }
};
