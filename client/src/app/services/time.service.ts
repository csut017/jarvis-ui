import { Injectable } from '@angular/core';
import * as moment from 'moment';

@Injectable({
  providedIn: 'root'
})
export class TimeService {

  constructor() { }

  getPartOfDay(): string {
    let now = moment();
    let dayPart = now.hour() < 12 ? 'morning' : (now.hour() < 17 ? ' afternoon' : 'evening');
    return dayPart;
  }

  getFriendlyTime(): string {
    let now = moment();
    let timePart = '';
    let minute = now.minute();
    switch (minute) {
      case 0:
        timePart = now.hour().toString();
        break;

      case 15:
        timePart = 'quarter past ' + now.hour().toString();
        break;

      case 30:
        timePart = 'half past ' + now.hour().toString();
        break;

      case 20:
        timePart = 'quarter to ' + (now.hour() + 1).toString();
        break;

      default:
        let minuteText = minute.toString();
        if (minute > 30) {
          timePart = minuteText + ' to ' + (now.hour() + 1).toString();
        } else {
          timePart = minuteText + ' past ' + now.hour().toString();
        }
        break;
    }

    return timePart;
  }
}
