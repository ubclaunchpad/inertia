import {
  TEST_DASHBOARD_ACTION,
} from '../actions/_constants';

const initialState = {
  testState: 'tree',
};

const Dashboard = (state = initialState, action) => {
  switch (action.type) {
    case TEST_DASHBOARD_ACTION: {
      return { ...state, testState: action.payload };
    }

    default: {
      return state;
    }
  }
};

export default Dashboard;
