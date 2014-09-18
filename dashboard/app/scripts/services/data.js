'use strict';

/**
 * @ngdoc service
 * @name dashboardApp.data
 * @description
 * # data
 * Service in the dashboardApp.
 */
angular.module('dashboardApp')
  .service('data', function ($http) {
    var dataAPI = {};

    dataAPI.get = function () {
      return $http({method: 'GET',
                    url: config.serviceURL});
    };

    return dataAPI;
  });
