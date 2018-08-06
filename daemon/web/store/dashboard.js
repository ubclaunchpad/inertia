import {
  GET_PROJECT_DETAILS_SUCCESS,
  GET_PROJECT_LOGS_SUCCESS,
  GET_CONTAINERS_SUCCESS,
} from '../actions/_constants';

const initialState = {
  project: {
    name: '',
    branch: '',
    commit: '',
    message: '',
    buildType: '',
  },
  logs: ['no logs'],
  containers: [],
};

const Dashboard = (state = initialState, action) => {
  switch (action.type) {
    case GET_PROJECT_DETAILS_SUCCESS: {
      return {
        ...state,
        project: action.payload.project,
      };
    }
    case GET_PROJECT_LOGS_SUCCESS: {
      return {
        ...state,
        logs: action.payload.logs,
      };
    }
    case GET_CONTAINERS_SUCCESS: {
      return {
        ...state,
        containers: action.payload.containers,
      };
    }

    default: {
      return state;
    }
  }
};

export default Dashboard;
