import {
  TEST_DASHBOARD_ACTION,
} from './_constants';


export const testAction = payload => (dispatch) => {
  // remove later
  console.log('dashboard action fired!');
  dispatch({
    type: TEST_DASHBOARD_ACTION,
    payload,
  });
};
