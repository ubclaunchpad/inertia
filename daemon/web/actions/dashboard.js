import {
  TEST_DASHBOARD_ACTION,
} from './_constants';


export const testAction = payload => (dispatch) => {
  // remove later
  dispatch({
    type: TEST_DASHBOARD_ACTION,
    payload,
  });
};
