import { Injectable } from '@angular/core';
import * as moment from 'moment';

@Injectable({
  providedIn: 'root'
})
export class TimeService {

  constructor() { }

  getPartOfDay(): string {
    let now = moment();
    return now.hour() < 12 ? 'morning' : (now.hour() < 18 ? ' afternoon' : 'evening');
  }

  getFriendlyDate(): string {
    let now = moment();
    return now.format('dddd, Do MMMM');
  }

  getFriendlyTime(): string {
    let now = moment();
    let minute = now.minute(),
        hour = now.hour();
    if (hour > 12) hour -=12;
    switch (minute) {
      case 0:
        return hour.toString();

      case 15:
        return 'quarter past ' + hour.toString();

      case 30:
        return 'half past ' + hour.toString();

      case 20:
        return 'quarter to ' + (hour + 1).toString();

      default:
        if (minute > 30) {
          return (60 - minute).toString() + ' to ' + (hour + 1).toString();
        }

        return minute.toString() + ' past ' + hour.toString();
    }
  }
}
