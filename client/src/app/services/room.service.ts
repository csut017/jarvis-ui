import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { LoggingService, Logger } from './logging.service';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { tap, catchError, map } from 'rxjs/operators';
import { List } from './common';

@Injectable({
  providedIn: 'root'
})
export class RoomService {

  constructor(private http: HttpClient,
    logging: LoggingService) {
    this.logger = logging.get('RoomService');
  }

  private logger: Logger;

  list(): Observable<Room[]> {
    const url = `${environment.apiURL}rooms`;
    this.logger.log(`Retrieving rooms`, url);
    return this.http.get<List<Room>>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved rooms', res)),
        catchError(this.logger.handleError<List<Room>>(`list()`)),
        map(res => res ? res.items : [])
      );
  }

  get(name: string): Observable<Room> {
    const uriName = encodeURIComponent(name);
    const url = `${environment.apiURL}rooms/${uriName}`;
    this.logger.log(`Retrieving room details for ${name}`, url);
    return this.http.get<Room>(url)
      .pipe(
        tap(res => this.logger.log('Retrieved room details', res)),
        catchError(this.logger.handleError<Room>(`get('${name}')`))
      );
  }
}

export interface Room {
  name: string;
  status: string;
}
