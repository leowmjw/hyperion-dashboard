'use strict';

/**
 * @ngdoc function
 * @name dashboardApp.controller:MainCtrl
 * @description
 * # MainCtrl
 * Controller of the dashboardApp
 */
angular.module('dashboardApp')
  .controller('MainCtrl', function ($scope, data, $log) {
    data.get()
      .success(function (data) {
        $scope.records = data.records;
      })
      .error(function (error) {
        $log.info(error);
      });
  });
