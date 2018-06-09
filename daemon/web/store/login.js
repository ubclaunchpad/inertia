import {
  TEST_LOGIN_ACTION,
} from '../actions/_constants';

const initialState = {
  testState: 'tree',
};

const Login = (state = initialState, action) => {
  switch (action.type) {
    case TEST_LOGIN_ACTION: {
      return { ...state, testState: action.payload };
    }

    default: {
      return state;
    }
  }
};

export default Login;
