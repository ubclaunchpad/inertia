import {
  GET_PROJECT_DETAILS_SUCCESS,
  GET_PROJECT_LOGS_SUCCESS,
  GET_PROJECT_LOGS_FAILURE,
  GET_CONTAINERS_SUCCESS,
} from './_constants';

const MOCK_DETAILS = {
  name: 'your-project-name',
  branch: 'master',
  commit: 'e51e133565bd0b0fda6caf69014dd2b2b24bfbaa',
  message: 'commit message goes here',
  buildType: 'docker-compose',
};

const MOCK_LOGS = [
  'log1asdasdasdasdasdasdasdssdasdasdssdasdasdssdasdasdssdasdasdsa',
  'log2asdasdasdasdsdassdasdasdssdasdasdssdasdasdssdasdasdsdsdasds',
  'log3dasdsdazxcxzsdasdasdssdasdasdssdasdasdssdasdasdsxxxxxxxxxx',
  'log4dasdsdasdsdasdasdssdasdasdssdasdasdssdasdasdsxzczxczxs',
  'log5dasdsdaasdsdasdasdssdasdasdssdasdasdssdasdasdsasdasdsds',
  'log6dasdsdaszsdasdasdssdasdasdssdasdasdssdasdasdsxczxczxczxcwqdqds',
  'log7dasdsdaxcsdasdasdssdasdasdssdasdasdssdasdasdszxczzxcsds',
];

const MOCK_CONTAINERS = [
  {
    name: '/inertia-deploy-test',
    status: 'ACTIVE',
    lastUpdated: '2018-01-01 00:00',
  },
  {
    name: '/docker-compose',
    status: 'ACTIVE',
    lastUpdated: '2018-01-01 00:00',
  },
];

function promiseState(p) {
  const t = {};

  return Promise.race([p, t])
    .then(v => (v === t ? 'pending' : ('fulfilled', () => 'rejected')));
}

export const handleGetProjectDetails = () => (dispatch) => {
  // TODO: put fetch request here
  dispatch({
    type: GET_PROJECT_DETAILS_SUCCESS,
    payload: { project: MOCK_DETAILS },
  });
};

export const handleGetLogs = ({ container }) => (dispatch) => {
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

export const handleGetContainers = () => (dispatch) => {
  // TODO: put fetch request here
  dispatch({
    type: GET_CONTAINERS_SUCCESS,
    payload: { containers: MOCK_CONTAINERS },
  });
};
